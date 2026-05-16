-- name: GetActiveMaintenanceWindow :one
SELECT *
FROM system_maintenance_windows
WHERE is_active = true
  AND (starts_at IS NULL OR starts_at <= now())
  AND (ends_at IS NULL OR ends_at >= now())
ORDER BY
  CASE mode WHEN 'global' THEN 1 WHEN 'read_only' THEN 2 ELSE 3 END,
  created_at DESC
LIMIT 1;

-- name: GetMaintenanceWindow :one
SELECT *
FROM system_maintenance_windows
WHERE id = sqlc.arg(id);

-- name: ListMaintenanceWindows :many
SELECT *
FROM system_maintenance_windows
ORDER BY created_at DESC
LIMIT LEAST(GREATEST(sqlc.arg(limit_count)::int, 1), 200)
OFFSET GREATEST(sqlc.arg(offset_count)::int, 0);

-- name: CreateMaintenanceWindow :one
INSERT INTO system_maintenance_windows (
  title, message, mode, affected_modules, starts_at, ends_at,
  is_active, allow_admin_bypass, bypass_roles, severity, created_by, updated_by
) VALUES (
  sqlc.arg(title),
  sqlc.arg(message),
  sqlc.arg(mode),
  sqlc.arg(affected_modules),
  sqlc.narg(starts_at),
  sqlc.narg(ends_at),
  sqlc.arg(is_active),
  sqlc.arg(allow_admin_bypass),
  sqlc.arg(bypass_roles),
  sqlc.arg(severity),
  sqlc.narg(actor_user_id),
  sqlc.narg(actor_user_id)
)
RETURNING *;

-- name: UpdateMaintenanceWindow :one
UPDATE system_maintenance_windows
SET title = sqlc.arg(title),
    message = sqlc.arg(message),
    mode = sqlc.arg(mode),
    affected_modules = sqlc.arg(affected_modules),
    starts_at = sqlc.narg(starts_at),
    ends_at = sqlc.narg(ends_at),
    allow_admin_bypass = sqlc.arg(allow_admin_bypass),
    bypass_roles = sqlc.arg(bypass_roles),
    severity = sqlc.arg(severity),
    updated_by = sqlc.narg(actor_user_id),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetMaintenanceWindowActive :one
UPDATE system_maintenance_windows
SET is_active = sqlc.arg(is_active),
    updated_by = sqlc.narg(actor_user_id),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CreateMaintenanceAuditLog :one
INSERT INTO system_maintenance_audit_logs (
  maintenance_id,
  actor_user_id,
  action,
  reason,
  metadata
) VALUES (
  sqlc.narg(maintenance_id),
  sqlc.narg(actor_user_id),
  sqlc.arg(action),
  sqlc.narg(reason),
  COALESCE(sqlc.arg(metadata)::jsonb, '{}'::jsonb)
)
RETURNING *;

-- name: ListMaintenanceAuditLogs :many
SELECT *
FROM system_maintenance_audit_logs
WHERE (sqlc.arg(action_filter)::text = '' OR action = sqlc.arg(action_filter)::text)
  AND (sqlc.narg(from_at)::timestamptz IS NULL OR created_at >= sqlc.narg(from_at)::timestamptz)
  AND (sqlc.narg(to_at)::timestamptz IS NULL OR created_at <= sqlc.narg(to_at)::timestamptz)
ORDER BY created_at DESC
LIMIT LEAST(GREATEST(sqlc.arg(limit_count)::int, 1), 200)
OFFSET GREATEST(sqlc.arg(offset_count)::int, 0);

-- name: GetMaintenanceDatabaseTime :one
SELECT now()::timestamptz AS db_time;

-- name: CountActiveCbtSessionsForMaintenance :one
SELECT COUNT(*)::bigint
FROM cbt_exam_sessions
WHERE status = 'active'
  AND scheduled_start <= now()
  AND scheduled_end >= now();
