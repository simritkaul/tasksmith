# AGENTS.md — TaskSmith

## What this project is

TaskSmith is a Go-based background job orchestration system, built as a learning-focused resume project. It is a restart — an earlier attempt stalled in 2025 with nothing usable carried over. This is scoped to be **finished**, not maximally ambitious.

## Architecture

- **Postgres** — owns job _definitions_ and _state_ (schedule, config, status, retry count, history).
- **QueueForge** (github.com/simritkaul/queueforge) — a separate, already-built Go project owning _dispatch_. Currently a **Go library**, not a network service: single in-process `sync.Mutex`-guarded `Queue`, WAL-backed, no server/client. TaskSmith imports it directly in Phase 0.
- **Worker pool** — goroutines pulling tasks, executing, reporting back to Postgres.
- **Retry engine** — on failure, re-enqueues via QueueForge with backoff, tracked in Postgres.
- **`/metrics` endpoint** — queue depth, success/failure rates, worker utilization.
- **Interface** — REST API + CLI client only.

## Critical design seam: the Dispatcher interface

QueueForge is never called directly from scheduler/worker code. All access goes through:

```go
type Dispatcher interface {
    Enqueue(payload string) (jobID string, err error)
    Dequeue() (*queue.Job, error)
    Ack(jobID string) error
}
```

Phase 0 has one implementation: a thin adapter around the in-process `queue.Queue`. This seam exists so a future networked/multi-process QueueForge can be swapped in later without touching scheduler or worker code. Do not bypass this interface anywhere, even for "quick" code.

## Classification — be precise, this is non-negotiable

Phase 0 is a **concurrent system built with distributed-systems primitives**, NOT a distributed system. Do not write, suggest, or generate code comments, docs, commit messages, or README language that calls this "distributed" or "production-grade."

- "Distributed" is earned only after: multiple independent worker _processes_ (not goroutines) run against a shared queue, with a worker-crash-mid-task test proving at-least-once delivery. This is an explicit, not-yet-started Phase 1 milestone.
- "Production-grade" is earned only after structured logging, graceful shutdown, health checks, and failure-injection tests actually exist. Good architecture is not the same thing as production-readiness.

## Explicit scope boundaries — do not suggest otherwise

- No Next.js, no web frontend, no UI framework of any kind. REST API + CLI only.
- No claiming "distributed" until the Phase 1 crash test is done and demonstrated.
- No claiming "production-grade" prematurely.
- Infra stays simple: Docker Compose (Postgres + TaskSmith) + basic CI (lint → test → build). Do not propose Kubernetes, service meshes, or other infra escalation for this project.
- No scope additions mid-build (new services, new queues, new databases) without the user explicitly deciding to add them first.

## How to help (this matters more than usual)

- The user is writing all the Go code themselves, by hand, to actually learn it — not having an agent generate it. Do not write full implementations unless explicitly asked to. Prefer: explain the concept, point out the bug, suggest the shape of a solution, ask what they've tried — not "here's the code."
- Be direct about mistakes, overscoping, or overclaiming — don't soften technical feedback to be encouraging. The user explicitly wants honesty over false positivity.
- If a suggestion conflicts with the classification/scope rules above, don't make the suggestion — or if asked directly, say clearly why it would be an overclaim or scope creep.
- The user's background: ~4 YOE, mainly React.js and .NET Core, learning Go specifically through this project and through a related project called CacheFlow. Explanations can assume solid general programming fluency but should not assume prior Go idiom knowledge.

## What "done" looks like for Phase 0

- Postgres schema for job definitions/state.
- Scheduler triggering jobs on a cron-like schedule.
- Worker pool (goroutines) executing via the Dispatcher interface.
- Retry engine with backoff, tracked in Postgres.
- `/metrics` endpoint.
- REST API + CLI client.
- Docker Compose (Postgres + TaskSmith).
- CI: lint → test → build.
- No claims beyond what is actually built and tested.
