-- max_attempts: column which tells the workers about the number of attempts they can make before the job goes into failed state
-- Not nullable, so we must explicitly pass 1 for non-retryable jobs
-- No index needed per current scale

-- backoff_base_seconds: column which tells the worker about the base backoff in seconds before that task can be marked pending again

ALTER TABLE job_definitions
ADD COLUMN IF NOT EXISTS max_attempts INT NOT NULL DEFAULT 3 CHECK (max_attempts > 0),
ADD COLUMN IF NOT EXISTS backoff_base_seconds INT NOT NULL DEFAULT 10 CHECK (backoff_base_seconds > 0);

-- To revert
-- ALTER TABLE job_definitions DROP COLUMN max_attempts, DROP COLUMN backoff_base_seconds;