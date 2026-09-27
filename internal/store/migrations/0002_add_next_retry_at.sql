-- next_retry_at: column which the poller queries against
-- Nullable as most runs will not be retried (happy flow)
-- No index for now as no need based on current scale
ALTER TABLE job_runs
ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;

-- To revert
-- ALTER TABLE job_runs DROP COLUMN next_retry_at;