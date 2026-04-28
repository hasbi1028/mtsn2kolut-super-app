-- name: ListSchedules :many
SELECT id, label, run_time, run_type, is_enabled, created_at, updated_at, last_enqueued_for_date
FROM schedules
ORDER BY run_type ASC, run_time ASC;

-- name: GetSchedule :one
SELECT id, label, run_time, run_type, is_enabled, created_at, updated_at, last_enqueued_for_date
FROM schedules
WHERE id = $1;

-- name: CreateSchedule :one
INSERT INTO schedules (label, run_time, run_type, is_enabled)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateScheduleByID :one
UPDATE schedules
SET label      = $2,
    run_time   = $3,
    is_enabled = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteScheduleByID :exec
DELETE FROM schedules WHERE id = $1;

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
