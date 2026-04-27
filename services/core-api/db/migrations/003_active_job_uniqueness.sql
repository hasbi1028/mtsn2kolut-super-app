-- Migration: 003_active_job_uniqueness
-- Purpose:
--   1. Rapikan duplicate active jobs lama jika ada
--   2. Pastikan hanya ada satu active job per employee + run_type

WITH ranked AS (
  SELECT
    id,
    ROW_NUMBER() OVER (
      PARTITION BY employee_id, run_type
      ORDER BY created_at ASC, id ASC
    ) AS rn
  FROM jobs
  WHERE status IN ('queued', 'running')
),
dupes AS (
  SELECT id
  FROM ranked
  WHERE rn > 1
)
UPDATE jobs
SET status = 'failed',
    error_message = CASE
      WHEN error_message = '' THEN 'Dibatalkan otomatis saat migrasi: duplicate active job'
      ELSE error_message
    END,
    next_retry_at = NULL,
    updated_at = NOW()
WHERE id IN (SELECT id FROM dupes);

CREATE UNIQUE INDEX IF NOT EXISTS uq_jobs_active_employee_run_type
  ON jobs (employee_id, run_type)
  WHERE status IN ('queued', 'running');
