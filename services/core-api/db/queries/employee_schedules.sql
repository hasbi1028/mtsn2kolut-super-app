-- name: ListEmployeeSchedules :many
SELECT id, employee_id, run_type, run_time, is_enabled, random_window_minutes,
       day_of_week, last_enqueued_for_date, created_at, updated_at
FROM employee_schedules
WHERE employee_id = $1
ORDER BY day_of_week ASC, run_type ASC;

-- name: UpsertEmployeeSchedule :one
INSERT INTO employee_schedules (employee_id, run_type, run_time, is_enabled, random_window_minutes, day_of_week)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (employee_id, run_type, day_of_week) DO UPDATE
  SET run_time              = EXCLUDED.run_time,
      is_enabled            = EXCLUDED.is_enabled,
      random_window_minutes = EXCLUDED.random_window_minutes,
      updated_at            = NOW()
RETURNING *;

-- name: DeleteEmployeeSchedule :exec
DELETE FROM employee_schedules WHERE id = $1 AND employee_id = $2;

-- name: ClaimDueEmployeeSchedules :many
WITH due AS (
  UPDATE employee_schedules es
  SET last_enqueued_for_date = $1,
      updated_at = NOW()
  FROM pusaka_accounts pa
  WHERE es.employee_id = pa.employee_id
    AND es.is_enabled = TRUE
    AND es.run_time <= $2
    AND es.day_of_week = EXTRACT(DOW FROM $1::DATE)::SMALLINT
    AND (es.last_enqueued_for_date IS NULL OR es.last_enqueued_for_date < $1)
    AND pa.is_enabled = TRUE
    AND pa.pusaka_username <> ''
    AND pa.pusaka_password <> ''
  RETURNING es.id,
            es.employee_id,
            es.run_type,
            es.run_time,
            es.is_enabled,
            es.random_window_minutes,
            es.day_of_week,
            es.last_enqueued_for_date,
            es.created_at,
            es.updated_at
)
SELECT * FROM due ORDER BY run_time ASC;

-- name: ResetEmployeeScheduleEnqueueState :exec
UPDATE employee_schedules
SET last_enqueued_for_date = NULL,
    updated_at = NOW()
WHERE id = $1 AND last_enqueued_for_date = $2;
