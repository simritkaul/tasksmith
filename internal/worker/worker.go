package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/simritkaul/tasksmith/internal/dispatcher"
	"github.com/simritkaul/tasksmith/internal/store"
)

type Worker struct {
	Dispatcher  dispatcher.Dispatcher
	Store       *store.Store
	Task        Task
	BackoffFunc BackoffFunc
}

// Initializes a new Worker with the provided dispatcher, store, task, and backoff function.
func New(d dispatcher.Dispatcher, s *store.Store, t Task, b BackoffFunc) *Worker {
	return &Worker{
		Dispatcher:  d,
		Store:       s,
		Task:        t,
		BackoffFunc: b,
	}
}

// Starts a worker
func (w *Worker) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		dequeued, err := w.processOneJob(ctx)
		if err != nil {
			log.Printf("worker: processing job: %v", err)
		}

		if !dequeued {
			timer := time.NewTimer(100 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}
}

// processOneJob dequeues the next job, loads its run configuration, and
// prepares the task execution for that job.
func (w *Worker) processOneJob(ctx context.Context) (bool, error) {
	job, err := w.Dispatcher.Dequeue()
	if errors.Is(err, dispatcher.ErrEmpty) {
		// The queue is empty, wait for a job
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("dequeueing a job: %w", err)
	}

	runID, err := uuid.Parse(job.Payload)
	if err != nil {
		return true, fmt.Errorf("parsing the job payload: %w", err)
	}

	rc, err := w.Store.GetRunConfig(ctx, runID)
	if err != nil {
		return true, fmt.Errorf("fetching the latest run config: %w", err)
	}

	var config map[string]any
	if err := json.Unmarshal(rc.Config, &config); err != nil {
		return true, fmt.Errorf("decoding run config: %w", err)
	}

	if err := w.Store.MarkRunStarted(ctx, runID); err != nil {
		return true, fmt.Errorf("mark run started: %w", err)
	}

	if err := w.Task(ctx, config); err != nil {
		// Task failed
		newAttemptCount := rc.AttemptCount + 1
		if newAttemptCount >= rc.MaxAttempts {
			if err := w.Store.MarkRunFailedFinal(ctx, runID, err.Error()); err != nil {
				return true, fmt.Errorf("mark run failed: %w", err)
			}

			if err := w.ack(job.ID); err != nil {
				return true, err
			}
		} else {
			backoffDuration := w.BackoffFunc(newAttemptCount, rc.BackoffBaseSeconds)
			if err := w.Store.MarkRunFailedRetrying(ctx, runID, time.Now().Add(backoffDuration), err.Error()); err != nil {
				return true, fmt.Errorf("mark run retry: %w", err)
			}

			if err := w.ack(job.ID); err != nil {
				return true, err
			}
		}
		return true, nil
	}

	if err := w.Store.MarkRunSucceeded(ctx, runID); err != nil {
		return true, fmt.Errorf("mark run succeeded: %w", err)
	}

	if err := w.ack(job.ID); err != nil {
		return true, err
	}

	return true, nil
}

func (w *Worker) ack(jobID string) error {
	if err := w.Dispatcher.Ack(jobID); err != nil {
		return fmt.Errorf("ack job: %w", err)
	}
	return nil
}
