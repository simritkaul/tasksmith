# Smoke Test

This directory contains a small end-to-end smoke test for the current Phase 0 path through TaskSmith's scheduler, dispatcher, QueueForge queue, and Postgres store. It is a standalone executable, not a Go test (`go test`). It does not start the scheduler loop or worker pool; it invokes scheduling once and simulates a worker inline.

## What It Exercises

Running the program performs these steps:

1. Opens `tasksmith.wal` in the current working directory and creates an in-process QueueForge queue and TaskSmith dispatcher.
2. Connects to the local Postgres database at `localhost:5433`, using the `tasksmith` database credentials from `docker-compose.yml`.
3. Inserts a throwaway `smoketest-job` definition with a sample JSON config.
4. Calls `ScheduleRun` once for the current time. This creates a pending row in `job_runs` and enqueues the run ID through the dispatcher.
5. Dequeues the queued job, parses its payload as a run ID, and fetches the run's config from Postgres.
6. Acknowledges the queue job and confirms that a subsequent dequeue reports an empty queue.

A successful run prints `Queue is empty as expected - smoke test passed!`. The program does not delete its Postgres rows or WAL file on exit. Re-running without cleanup will fail when it tries to insert another job definition with the unique name `smoketest-job`.

## Run

From the repository root, start Postgres and run the executable:

```sh
docker compose up -d postgres
go run ./cmd/smoketest
```

Run it from the repository root because `tasksmith.wal` is a relative path. The database connection is hard-coded for the Compose port mapping (`localhost:5433`).

## Cleanup

After the process exits, clear the smoke-test rows and remove the local WAL file. The `TRUNCATE` command below empties both TaskSmith tables, so use it only when it is safe to remove **all** job definitions and runs in this database; it is intended for the local development database, not a database containing data you need.

```sh
docker compose exec -T postgres psql -U tasksmith -d tasksmith \
  -c 'TRUNCATE TABLE job_runs, job_definitions;'
rm -f tasksmith.wal
```

The WAL file is created in the directory from which the program was run. With the documented command, that is the repository root. Stop the smoke test before removing it so the queue has closed the file. If the test is run from another directory, remove the corresponding `tasksmith.wal` in that directory instead.
