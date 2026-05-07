-- name: ListJobs :many
SELECT j.id, j.employee_id, e.nama AS employee_nama, COALESCE(e.nip, '')::text AS employee_nip,
       j.run_type, j.status, j.error_message,
       j.claimed_by, j.claimed_at, j.attempts, j.max_attempts,
       j.next_retry_at, j.created_at, j.updated_at
FROM jobs j
JOIN employees e ON e.id = j.employee_id
ORDER BY j.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListJobsByStatus :many
SELECT j.id, j.employee_id, e.nama AS employee_nama, COALESCE(e.nip, '')::text AS employee_nip,
       j.run_type, j.status, j.error_message,
       j.claimed_by, j.claimed_at, j.attempts, j.max_attempts,
       j.next_retry_at, j.created_at, j.updated_at
FROM jobs j
JOIN employees e ON e.id = j.employee_id
WHERE j.status = $1
ORDER BY j.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountJobs :one
SELECT COUNT(*) FROM jobs;

-- name: CountJobsByStatus :one
SELECT COUNT(*) FROM jobs WHERE status = $1;

-- name: CreateJobIfAbsent :one
INSERT INTO jobs (id, employee_id, run_type, status, max_attempts, not_before)
VALUES (gen_random_uuid(), $1, $2, 'queued', $3, $4)
ON CONFLICT (employee_id, run_type)
  WHERE status IN ('queued', 'running')
DO NOTHING
RETURNING *;

-- name: ClaimJob :one
WITH candidate AS (
  SELECT j.id
  FROM jobs j
  JOIN pusaka_accounts pa ON pa.employee_id = j.employee_id
  WHERE (
      (j.status = 'queued' AND (j.not_before IS NULL OR j.not_before <= NOW()))
      OR (j.status = 'failed' AND j.attempts < j.max_attempts AND j.next_retry_at <= NOW())
    )
    AND pa.is_enabled = TRUE
    AND pa.pusaka_username <> ''
    AND pa.pusaka_password <> ''
  ORDER BY j.created_at ASC
  FOR UPDATE SKIP LOCKED
  LIMIT 1
),
updated AS (
  UPDATE jobs
  SET status     = 'running',
      claimed_by = $1,
      claimed_at = NOW(),
      attempts   = attempts + 1,
      updated_at = NOW()
  FROM candidate
  WHERE jobs.id = candidate.id
  RETURNING jobs.id, jobs.employee_id, jobs.run_type, jobs.attempts, jobs.max_attempts
)
SELECT u.id, u.employee_id, u.run_type, u.attempts, u.max_attempts,
       pa.pusaka_username, pa.pusaka_password
FROM updated u
JOIN pusaka_accounts pa ON pa.employee_id = u.employee_id;

-- name: RecoverStaleRunningJobs :one
WITH updated AS (
  UPDATE jobs
  SET status        = 'failed',
      error_message = 'Job running melewati batas waktu pemulihan; akan dicoba ulang jika jatah percobaan masih ada.',
      claimed_by    = '',
      claimed_at    = NULL,
      next_retry_at = CASE
        WHEN attempts < max_attempts THEN NOW()
        ELSE NULL
      END,
      updated_at    = NOW()
  WHERE status = 'running'
    AND claimed_at IS NOT NULL
    AND claimed_at < NOW() - (sqlc.arg(stale_after_seconds)::INT || ' seconds')::INTERVAL
  RETURNING 1
)
SELECT COUNT(*)::BIGINT FROM updated;

-- name: CompleteJob :execrows
UPDATE jobs
SET status        = 'success',
    error_message = '',
    claimed_by    = '',
    claimed_at    = NULL,
    updated_at    = NOW()
WHERE id = $1
  AND status = 'running'
  AND claimed_by = $2;

-- name: FailJob :execrows
UPDATE jobs
SET status        = 'failed',
    error_message = $2,
    claimed_by    = '',
    claimed_at    = NULL,
    next_retry_at = CASE
      WHEN attempts < max_attempts THEN NOW() + ($4 || ' seconds')::INTERVAL
      ELSE NULL
    END,
    updated_at    = NOW()
WHERE id = $1
  AND status = 'running'
  AND claimed_by = $3;

-- name: GetJob :one
SELECT id, employee_id, run_type, status, error_message, claimed_by, claimed_at,
       attempts, max_attempts, next_retry_at, created_at, updated_at
FROM jobs WHERE id = $1;

-- name: CancelEmployeeJobs :one
WITH updated AS (
  UPDATE jobs
  SET status = 'failed', error_message = 'Dibatalkan manual', next_retry_at = NULL, updated_at = NOW()
  WHERE employee_id = $1 AND (status = 'queued' OR status = 'running')
  RETURNING 1
)
SELECT COUNT(*) FROM updated;

-- name: CancelAllJobs :one
WITH updated AS (
  UPDATE jobs
  SET status = 'failed', error_message = 'Dibatalkan manual', next_retry_at = NULL, updated_at = NOW()
  WHERE status = 'queued' OR status = 'running'
  RETURNING 1
)
SELECT COUNT(*) FROM updated;

-- name: GetJobStats :one
SELECT
  COUNT(*) FILTER (WHERE status = 'queued')  AS queued,
  COUNT(*) FILTER (WHERE status = 'running') AS running,
  COUNT(*) FILTER (WHERE status = 'success') AS success,
  COUNT(*) FILTER (WHERE status = 'failed')  AS failed
FROM jobs;
