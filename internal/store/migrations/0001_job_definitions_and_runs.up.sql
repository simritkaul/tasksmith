-- job_definitions: one row per logical job (schedule/config). Low churn.
CREATE TABLE job_definitions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT NOT NULL UNIQUE,
    cron_schedule  TEXT NOT NULL,
    config         JSONB NOT NULL DEFAULT '{}',
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- job_runs: one row per trigger occurrence. Doubles as the history/log table.
-- attempt_count/last_error are scoped to this run only (reset per trigger,
-- not cumulative across the job definition's lifetime).
CREATE TABLE job_runs (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_definition_id  UUID NOT NULL REFERENCES job_definitions(id),
    status             TEXT NOT NULL DEFAULT 'pending',
    scheduled_at       TIMESTAMPTZ NOT NULL,
    started_at         TIMESTAMPTZ,
    completed_at       TIMESTAMPTZ,
    attempt_count      INT NOT NULL DEFAULT 0,
    last_error         TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT job_runs_status_check
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed'))
);

CREATE INDEX idx_job_runs_job_definition_id ON job_runs(job_definition_id);
CREATE INDEX idx_job_runs_status ON job_runs(status);

-- Reconciliation query (not automated yet — Phase 0 sketch only):
-- finds job_runs rows stuck at 'pending' well past their scheduled_at,
-- meaning the Postgres insert committed but the QueueForge enqueue call
-- either failed or never happened. These are candidates for re-enqueueing.
--
-- SELECT id, job_definition_id, scheduled_at
-- FROM job_runs
-- WHERE status = 'pending'
--   AND scheduled_at < now() - INTERVAL '5 minutes';