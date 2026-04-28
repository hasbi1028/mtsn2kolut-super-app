-- name: ListEmployeeSchedules :many
SELECT id, employee_id, run_type, run_time, is_enabled, last_enqueued_for_date, created_at, updated_at
FROM employee_schedules
WHERE employee_id = $1
ORDER BY run_type ASC;

-- name: UpsertEmployeeSchedule :one
INSERT INTO employee_schedules (employee_id, run_type, run_time, is_enabled)
VALUES ($1, $2, $3, $4)
ON CONFLICT (employee_id, run_type) DO UPDATE
  SET run_time   = EXCLUDED.run_time,
      is_enabled = EXCLUDED.is_enabled,
      updated_at = NOW()
RETURNING *;

-- name: DeleteEmployeeSchedule :exec
DELETE FROM employee_schedules WHERE id = $1 AND employee_id = $2;

-- name: ClaimDueEmployeeSchedules :many
WITH due AS (
  UPDATE employee_schedules
  SET last_enqueued_for_date = $1,
      updated_at = NOW()
  WHERE is_enabled = TRUE
    AND run_time <= $2
    AND (last_enqueued_for_date IS NULL OR last_enqueued_for_date < $1)
  RETURNING *
)
SELECT * FROM due ORDER BY run_time ASC;

-- name: ResetEmployeeScheduleEnqueueState :exec
UPDATE employee_schedules
SET last_enqueued_for_date = NULL,
    updated_at = NOW()
WHERE id = $1 AND last_enqueued_for_date = $2;
