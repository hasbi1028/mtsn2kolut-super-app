-- name: CreateInternalAnalyticsEvent :one
INSERT INTO internal_analytics_events (
  event_name,
  event_group,
  source_surface,
  actor_user_id,
  actor_role,
  permission_code,
  route_group,
  module,
  result,
  status_code_class,
  duration_bucket,
  metadata,
  retention_expires_at
)
VALUES (
  sqlc.arg(event_name),
  sqlc.arg(event_group),
  sqlc.arg(source_surface),
  sqlc.narg(actor_user_id),
  sqlc.narg(actor_role),
  sqlc.narg(permission_code),
  sqlc.narg(route_group),
  sqlc.narg(module),
  sqlc.narg(result),
  sqlc.narg(status_code_class),
  sqlc.narg(duration_bucket),
  COALESCE(sqlc.arg(metadata)::jsonb, '{}'::jsonb),
  sqlc.arg(retention_expires_at)
)
RETURNING id, event_name, event_group, occurred_at, source_surface, actor_user_id, actor_role, permission_code, route_group, module, result, status_code_class, duration_bucket, metadata, retention_expires_at, created_at;

-- name: GetInternalAnalyticsEvent :one
SELECT id, event_name, event_group, occurred_at, source_surface, actor_user_id, actor_role, permission_code, route_group, module, result, status_code_class, duration_bucket, metadata, retention_expires_at, created_at
FROM internal_analytics_events
WHERE id = sqlc.arg(id);

-- name: ListInternalAnalyticsEventsForRollup :many
SELECT id, event_name, event_group, occurred_at, source_surface, actor_user_id, actor_role, permission_code, route_group, module, result, status_code_class, duration_bucket, metadata, retention_expires_at, created_at
FROM internal_analytics_events
WHERE occurred_at >= sqlc.arg(start_at)
  AND occurred_at < sqlc.arg(end_at)
ORDER BY occurred_at, id
LIMIT LEAST(GREATEST(sqlc.arg(limit_count)::INT, 1), 10000);

-- name: DeleteExpiredInternalAnalyticsEvents :execrows
DELETE FROM internal_analytics_events
WHERE retention_expires_at <= sqlc.arg(cutoff_at);

-- name: UpsertInternalAnalyticsDailyAggregate :one
INSERT INTO internal_analytics_daily_aggregates (
  aggregate_date,
  event_group,
  event_name,
  source_surface,
  role,
  result,
  count,
  total_duration_ms,
  metadata
)
VALUES (
  sqlc.arg(aggregate_date),
  sqlc.arg(event_group),
  sqlc.arg(event_name),
  sqlc.arg(source_surface),
  sqlc.arg(role),
  sqlc.arg(result),
  sqlc.arg(event_count),
  sqlc.narg(total_duration_ms),
  COALESCE(sqlc.arg(metadata)::jsonb, '{}'::jsonb)
)
ON CONFLICT (aggregate_date, event_group, event_name, source_surface, role, result)
DO UPDATE SET
  count = EXCLUDED.count,
  total_duration_ms = EXCLUDED.total_duration_ms,
  metadata = EXCLUDED.metadata,
  updated_at = NOW()
RETURNING aggregate_date, event_group, event_name, source_surface, role, result, count, total_duration_ms, metadata, created_at, updated_at;

-- name: ListInternalAnalyticsDailyAggregates :many
SELECT aggregate_date, event_group, event_name, source_surface, role, result, count, total_duration_ms, metadata, created_at, updated_at
FROM internal_analytics_daily_aggregates
WHERE aggregate_date >= sqlc.arg(start_date)
  AND aggregate_date <= sqlc.arg(end_date)
  AND (sqlc.arg(event_group)::TEXT = '' OR event_group = sqlc.arg(event_group)::TEXT)
  AND (sqlc.arg(event_name)::TEXT = '' OR event_name = sqlc.arg(event_name)::TEXT)
  AND (sqlc.arg(source_surface)::TEXT = '' OR source_surface = sqlc.arg(source_surface)::TEXT)
  AND (sqlc.arg(role)::TEXT = '' OR role = sqlc.arg(role)::TEXT)
  AND (sqlc.arg(result)::TEXT = '' OR result = sqlc.arg(result)::TEXT)
ORDER BY aggregate_date DESC, event_group, event_name, source_surface, role, result
LIMIT LEAST(GREATEST(sqlc.arg(limit_count)::INT, 1), 10000)
OFFSET GREATEST(sqlc.arg(offset_count)::INT, 0);
