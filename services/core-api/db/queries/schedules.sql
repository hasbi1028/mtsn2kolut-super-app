-- name: ListSchedules :many
SELECT id, label, run_time, run_type, is_enabled, created_at, updated_at, last_enqueued_for_date
FROM schedules
ORDER BY run_type ASC;

-- name: GetSchedule :one
SELECT id, label, run_time, run_type, is_enabled, created_at, updated_at, last_enqueued_for_date
FROM schedules
WHERE id = $1;

-- name: UpsertSchedule :one
INSERT INTO schedules (id, label, run_time, run_type, is_enabled)
VALUES (gen_random_uuid(), $1, $2, $3, $4)
ON CONFLICT (run_type) DO UPDATE
  SET label      = EXCLUDED.label,
      run_time   = EXCLUDED.run_time,
      is_enabled = EXCLUDED.is_enabled,
      updated_at = NOW()
RETURNING *;

-- name: ClaimDueSchedules :many
WITH due AS (
  UPDATE schedules
  SET last_enqueued_for_date = $1,
      updated_at = NOW()
  WHERE is_enabled = TRUE
    AND run_time <= $2
    AND (last_enqueued_for_date IS NULL OR last_enqueued_for_date < $1)
  RETURNING *
)
SELECT id, label, run_time, run_type, is_enabled, created_at, updated_at, last_enqueued_for_date
FROM due
ORDER BY run_time ASC;

-- name: ResetScheduleEnqueueState :exec
UPDATE schedules
SET last_enqueued_for_date = NULL,
    updated_at = NOW()
WHERE id = $1 AND last_enqueued_for_date = $2;
