package store

import (
	"context"
	"database/sql"
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