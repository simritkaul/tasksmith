package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/simritkaul/tasksmith/internal/dispatcher"
	"github.com/simritkaul/tasksmith/internal/store"
)

type Scheduler struct {
	Store *store.Store;
	Dispatcher dispatcher.Dispatcher;
}

func New(s *store.Store, d dispatcher.Dispatcher) *Scheduler {
	return &Scheduler{
		Store: s,
		Dispatcher: d,
	};
}

/*
ScheduleRun creates a pending job_runs row, then enqueues it into the dispatcher.
These are two seperate non-atomic steps (a Postgres trasaction can't span QueueForge).
If `Enqueue` fails, the row is left 'pending' - recoverable via reconcilation, not silently lost.
*/
func (sch *Scheduler) ScheduleRun(ctx context.Context, jobDefId uuid.UUID, scheduledAt time.Time) (store.JobRun, error) {
	run, err := sch.Store.CreateJobRun(ctx, jobDefId, scheduledAt);
	if err != nil {
		return store.JobRun{}, fmt.Errorf("create job run: %w", err);
	}

	if _, err := sch.Dispatcher.Enqueue(run.ID.String()); err != nil {
		return run, fmt.Errorf("enqueue job run %s: %w", run.ID, err);
	}

	return run, nil;
}