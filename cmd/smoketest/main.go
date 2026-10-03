package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/simritkaul/queueforge/queue"
	"github.com/simritkaul/tasksmith/internal/dispatcher"
	"github.com/simritkaul/tasksmith/internal/scheduler"
	"github.com/simritkaul/tasksmith/internal/store"
)

func main() {
	ctx := context.Background()

	// QueueForge / Dispatcher setup
	wal, err := queue.OpenWAL("tasksmith.wal")
	if err != nil {
		log.Fatal(err)
	}

	defer wal.Close()

	q, err := queue.NewQueue(wal)
	if err != nil {
		log.Fatal(err)
	}

	disp := dispatcher.NewInprocessDispatcher(q)

	// Postgres setup
	dsn := "postgres://tasksmith:tasksmith@localhost:5433/tasksmith?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("could not reach postgres: ", err)
	}

	fmt.Println("Connected to Postgres")

	st := store.New(db)
	sch := scheduler.New(st, disp)

	// Throwaway Job Definition
	cfg := json.RawMessage(`{"message": "hello from smoke test"}`)

	jobDef, err := st.CreateJobDefinition(ctx, "smoketest-job", "* * * * *", cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created Job Definition: ", jobDef.ID)

	// Scheduler create pending run + enqueue
	run, err := sch.ScheduleRun(ctx, jobDef.ID, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Scheduled run: ", run.ID, "status: ", run.Status)

	// Simulated worker: dequeue, resolve run_id, fetch fresh config
	job, err := disp.Dequeue()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Worker dequeued job, payload (run_id): ", job.Payload)

	runID, err := uuid.Parse(job.Payload)
	if err != nil {
		log.Fatal("payload was not a valid run ID: ", err)
	}

	rc, err := st.GetRunConfig(ctx, runID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Worker fetched fresh config: %s\n", rc.Config)

	if err := disp.Ack(job.ID); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Acked Job: ", job.ID)

	// Confirm queue is now empty
	_, err = disp.Dequeue()
	if errors.Is(err, dispatcher.ErrEmpty) {
		fmt.Println("Queue is empty as expected - smoke test passed!")
	} else if err != nil {
		log.Fatal(err)
	} else {
		log.Fatal("expected empty queue but found a job")
	}
}
