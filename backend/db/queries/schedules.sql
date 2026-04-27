-- name: ListSchedules :many
SELECT id, label, run_time, run_type, is_enabled, created_at, updated_at
FROM schedules
ORDER BY run_type ASC;

-- name: GetSchedule :one
SELECT id, label, run_time, run_type, is_enabled, created_at, updated_at
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
