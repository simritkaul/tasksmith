package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/simritkaul/queueforge/queue"
	"github.com/simritkaul/tasksmith/internal/dispatcher"
	"github.com/simritkaul/tasksmith/internal/store"
	"github.com/simritkaul/tasksmith/internal/worker"
)

func main() {
	ctx := context.Background()

	db := mustOpenDB(ctx)
	defer db.Close()

	// st := store.New(db)
}

// Fresh WAL + queue per scenario, so scenarios can't see each other's messages.
func newDispatcher() (dispatcher.Dispatcher, func()) {
	dir, err := os.MkdirTemp("", "tasksmith-smoke-*")
	if err != nil {
		log.Fatal(err)
	}

	wal, err := queue.OpenWAL(filepath.Join(dir, "q.wal"))
	if err != nil {
		log.Fatal(err)
	}

	q, err := queue.NewQueue(wal)
	if err != nil {
		log.Fatal(err)
	}

	return dispatcher.NewInprocessDispatcher(q), func() {
		wal.Close()
		os.RemoveAll(dir)
	}
}

// Postgres setup
func mustOpenDB(ctx context.Context) *sql.DB {
	dsn := "postgres://tasksmith:tasksmith@localhost:5433/tasksmith?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("could not reach postgres: ", err)
	}

	fmt.Println("Connected to Postgres")

	return db
}

// Starts the real worker; the returned func stops it and WAITS for Run to return.
func startWorker(parent context.Context, st *store.Store, disp dispatcher.Dispatcher) func() {
	ctx, cancel := context.WithCancel(parent)
	done := make(chan struct{})

	w := worker.New(disp, st, worker.PlaceholderTask, worker.FixedBackoff)

	go func() {
		defer close(done)
		w.Run(ctx)
	}()

	return func() {
		cancel()
		<-done
	}
}
