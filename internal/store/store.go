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
};

func (s *Store) GetRunConfig(ctx context.Context, runId uuid.UUID) (RunWithConfig, error) {
	var rc RunWithConfig;
	err := s.db.QueryRowContext(ctx,
		`SELECT jr.id, jr.job_definition_id, jd.config
		 FROM job_runs jr
		 JOIN job_definitions jd 
		 ON jr.job_definition_id = jd.id
		 WHERE jr.id = $1`,
		 runId,
	).Scan(&rc.RunID, &rc.JobDefinitionID, &rc.Config);
	if err != nil {
		return RunWithConfig{}, fmt.Errorf("get run config: %w", err);
	}

	return rc, nil;
}