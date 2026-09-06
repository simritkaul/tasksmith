# TaskSmith

A Go-based background job orchestration system — schedule tasks, dispatch
them through a durable queue, execute them via a worker pool, and track
retries and history in Postgres.

> **Status: Phase 0, early build.** This project is scoped to be finished,
> not maximally ambitious. See [Classification](#classification--be-precise)
> below before assuming more than what's actually built.

## What it does

- **Postgres** owns job _definitions_ and _state_ — schedule (cron-like),
  config, status, retry count, history.
- **[QueueForge](https://github.com/simritkaul/queueforge)** — a
  WAL-backed durable queue, built as a separate project — owns _dispatch_.
  In Phase 0 it's used as an in-process Go library.
- **Worker pool** pulls tasks, executes them, and reports success/failure
  back to Postgres.
- **Retry engine** re-enqueues failed tasks with backoff, tracked in
  Postgres.
- **`/metrics`** exposes queue depth, success/failure rates, and worker
  utilization.
- **Interface**: REST API + CLI client. No web frontend.

## Classification — be precise

Phase 0 is a **concurrent system, built with distributed-systems
primitives** — it is **not** a distributed system yet, and this README
will not call it one.

It becomes distributed only once multiple independent worker _processes_
(not goroutines) run against a shared queue, with a demonstrated
crash-mid-task test proving at-least-once delivery. That's a named,
not-yet-started **Phase 1** milestone — not implied by anything below.

Likewise, "production-grade" is not claimed until structured logging,
graceful shutdown, health checks, and failure-injection tests actually
exist. Good architecture and production-readiness are different things.

## Architecture

```
Scheduler ──enqueue──▶ Dispatcher ──▶ QueueForge (WAL-backed queue)
                                            │
                                        dequeue
                                            ▼
                                      Worker Pool ──report──▶ Postgres
                                            │                    │
                                        on failure          job state,
                                            ▼                  retries,
                                      Retry Engine ──────────▶ history
```

TaskSmith never calls QueueForge directly — all access goes through a
`Dispatcher` interface. Phase 0's implementation wraps the in-process
QueueForge library; a future networked/multi-process QueueForge could be
swapped in later without touching scheduler or worker code.

## Tech stack

- Go
- PostgreSQL
- QueueForge (own WAL-based queue library)
- Docker Compose
- CI: lint → test → build

## Roadmap

- [ ] Postgres schema for job definitions/state
- [ ] Scheduler (cron-like triggering)
- [ ] Dispatcher interface + Phase 0 (in-process) adapter
- [ ] Worker pool
- [ ] Retry engine with backoff
- [ ] `/metrics` endpoint
- [ ] REST API + CLI client
- [ ] Docker Compose (Postgres + TaskSmith)
- [ ] CI pipeline (lint → test → build)
- [ ] **Phase 1** (separate milestone, not started): networked QueueForge,
      multiple worker processes, crash-mid-task correctness test

## Why this project exists

Built to genuinely understand — not just namedrop — the difference between
a concurrent system and a distributed one; what "correct under partial
failure" requires (idempotency, at-least-once vs. exactly-once, what a WAL
actually buys you); and how to compose independently-built systems
(QueueForge as a dependency) instead of rebuilding everything from scratch
each time.

## License

TBD
