package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB;
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

type JobRun struct {
	ID	uuid.UUID;
	JobDefinitionID	uuid.UUID;
	Status string;
	ScheduledAt time.Time;
	AttemptCount int;
}

// CreateJobRun inserts a new job_runs row with status = 'pending'
func (s *Store) CreateJobRun(ctx context.Context, jobDefId uuid.UUID, scheduledAt time.Time) (JobRun, error) {
	id := uuid.New();
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO job_runs (id, job_definition_id, status, scheduled_at) 
			VALUES ($1, $2, 'pending', $3)`,
			id, jobDefId, scheduledAt,
	)

	if err != nil {
		return JobRun{}, fmt.Errorf("insert job run: %w", err);
	}

	return JobRun{
		ID: id,
		JobDefinitionID: jobDefId,
		Status: "pending",
		ScheduledAt: scheduledAt,
	}, nil;
}

type JobDefinition struct {
	ID uuid.UUID;
	Name string;
	CronSchedule string;
	Config json.RawMessage;
	Enabled bool;
}

// CreateJobDefinition inserts a new job_definitions row
func (s *Store) CreateJobDefinition(ctx context.Context, name string, cronSchedule string, config json.RawMessage) (JobDefinition, error) {
	id := uuid.New();
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO job_definitions(id, name, cron_schedule, config)
		 VALUES ($1, $2, $3, $4)`,
		 id, name, cronSchedule, config,
	)

	if err != nil {
		return JobDefinition{}, fmt.Errorf("insert job definition: %w", err);
	}

	return JobDefinition{
		ID: id,
		Name: name,
		CronSchedule: cronSchedule,
		Config: config,
		Enabled: true, // matches column default
	}, nil;
}

type RunWithConfig struct {
	RunID uuid.UUID;
	JobDefinitionID uuid.UUID;
	Config json.RawMessage;
	MaxAttempts int;
};

// GetRunConfig gets the latest job details for a given runID
func (s *Store) GetRunConfig(ctx context.Context, runID uuid.UUID) (RunWithConfig, error) {
	var rc RunWithConfig;
	err := s.db.QueryRowContext(ctx,
		`SELECT jr.id, jr.job_definition_id, jd.config, jd.max_attempts
		 FROM job_runs jr
		 JOIN job_definitions jd 
		 ON jr.job_definition_id = jd.id
		 WHERE jr.id = $1`,
		 runID,
	).Scan(&rc.RunID, &rc.JobDefinitionID, &rc.Config, &rc.MaxAttempts);
	if err != nil {
		return RunWithConfig{}, fmt.Errorf("get run config: %w", err);
	}

	return rc, nil;
}

// MarkRunStarted marks the job with the given runID as running
func (s *Store) MarkRunStarted(ctx context.Context, runID uuid.UUID) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE job_runs
		 SET status = 'running', 
		 started_at = now(), 
		 attempt_count = attempt_count + 1, 
		 next_retry_at = NULL
		 WHERE id = $1`,
		runID,
	)

	if err != nil {
		return fmt.Errorf("mark run started: %w", err);
	}

	rows, err := res.RowsAffected();

	if err != nil {
		return fmt.Errorf("mark run started: checking rows affected: %w", err);
	}

	if rows != 1 {
		return fmt.Errorf("mark run started: expected 1 row affected, got %d (run %s not found?)", rows, runID);
	}

	return nil;
}

// MarkRunSucceeded marks the job with the given runID as succeeded
func (s *Store) MarkRunSucceeded(ctx context.Context, runID uuid.UUID) error {
	res, err := s.db.ExecContext(ctx, 
		`UPDATE job_runs 
		 SET status = 'succeeded', completed_at = now()
		 WHERE id = $1`,
		 runID,
	)

	if err != nil {
		return fmt.Errorf("mark run succeeded: %w", err);
	}

	rows, err := res.RowsAffected();

	if err != nil {
		return fmt.Errorf("mark run succeeded: checking rows affected: %w", err);
	}

	if rows != 1 {
		return fmt.Errorf("mark run succeeded: expected 1 row affected, got %d (run %s not found?)", rows, runID);
	}

	return nil;
}

// MarkRunFailedRetrying marks the job with the given runID as pending and sets the next run at for the retry
func (s *Store) MarkRunFailedRetrying(ctx context.Context, runID uuid.UUID, nextRunAt time.Time, errMsg string) error {
	res, err := s.db.ExecContext(ctx, 
		`UPDATE job_runs 
		 SET status = 'pending', next_retry_at = $1, last_error = $2
		 WHERE id = $3`,
		 nextRunAt,
		 errMsg,
		 runID,
	)

	if err != nil {
		return fmt.Errorf("mark run failed retrying: %w", err);
	}

	rows, err := res.RowsAffected();

	if err != nil {
		return fmt.Errorf("mark run failed retrying: checking rows affected: %w", err);
	}

	if rows != 1 {
		return fmt.Errorf("mark run failed retrying: expected 1 row affected, got %d (run %s not found?)", rows, runID);
	}

	return nil;
}

// MarkRunFailedFinal marks the job with the given runID as failed
func (s *Store) MarkRunFailedFinal(ctx context.Context, runID uuid.UUID, errMsg string) error {
	res, err := s.db.ExecContext(ctx, 
		`UPDATE job_runs 
		 SET status = 'failed', 
		 completed_at = now(),
		 next_retry_at = NULL, 
		 last_error = $1
		 WHERE id = $2`,
		 errMsg,
		 runID,
	)

	if err != nil {
		return fmt.Errorf("mark run failed final: %w", err);
	}

	rows, err := res.RowsAffected();

	if err != nil {
		return fmt.Errorf("mark run failed final: checking rows affected: %w", err);
	}

	if rows != 1 {
		return fmt.Errorf("mark run failed final: expected 1 row affected, got %d (run %s not found?)", rows, runID);
	}

	return nil;
}