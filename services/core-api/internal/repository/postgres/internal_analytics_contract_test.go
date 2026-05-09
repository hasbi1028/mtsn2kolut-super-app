package db

import (
	"os"
	"strings"
	"testing"
)

const internalAnalyticsMigrationPath = "../../../db/migrations/083_internal_analytics_schema.sql"
const internalAnalyticsQueriesPath = "../../../db/queries/internal_analytics.sql"

func readLowerFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.ToLower(string(content))
}

func TestInternalAnalyticsPhase1MigrationContract(t *testing.T) {
	sql := readLowerFile(t, internalAnalyticsMigrationPath)

	required := []string{
		"create table if not exists internal_analytics_events",
		"id uuid primary key default gen_random_uuid()",
		"event_name text not null",
		"event_group text not null",
		"occurred_at timestamptz not null default now()",
		"source_surface text not null",
		"actor_user_id uuid",
		"actor_role text",
		"permission_code text",
		"route_group text",
		"module text",
		"result text",
		"status_code_class text",
		"duration_bucket text",
		"metadata jsonb not null default '{}'::jsonb",
		"retention_expires_at timestamptz not null",
		"created_at timestamptz not null default now()",
		"constraint chk_internal_analytics_events_group",
		"event_group in ('public', 'auth', 'dashboard', 'bank_soal', 'asesmen', 'pusaka', 'users_rbac', 'security')",
		"constraint chk_internal_analytics_events_name",
		"event_name ~ '^[a-z][a-z0-9_]*\\.[a-z][a-z0-9_]*$'",
		"jsonb_typeof(metadata) = 'object'",
		"create index if not exists idx_internal_analytics_events_occurred_at",
		"create index if not exists idx_internal_analytics_events_retention_expires_at",
		"create index if not exists idx_internal_analytics_events_group_occurred_at",
		"create index if not exists idx_internal_analytics_events_name_occurred_at",
		"create table if not exists internal_analytics_daily_aggregates",
		"aggregate_date date not null",
		"role text not null default ''",
		"count bigint not null default 0",
		"total_duration_ms bigint",
		"updated_at timestamptz not null default now()",
		"constraint uq_internal_analytics_daily_aggregates_key unique",
	}

	for _, needle := range required {
		if !strings.Contains(sql, needle) {
			t.Fatalf("internal analytics migration missing %q", needle)
		}
	}
}

func TestInternalAnalyticsPhase1KeepsSensitiveFieldsOutOfSchema(t *testing.T) {
	sql := readLowerFile(t, internalAnalyticsMigrationPath)

	forbiddenColumns := []string{
		"raw_ip",
		"raw_user_agent",
		"password",
		"token",
		"cookie",
		"authorization",
		"secret",
		"api_key",
		"nik",
		"nip",
		"nisn",
		"device_fingerprint",
		"credential_pusaka",
	}

	for _, forbidden := range forbiddenColumns {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("internal analytics migration must not contain sensitive schema string %q", forbidden)
		}
	}
}

func TestInternalAnalyticsPhase1DoesNotTouchAuditLogs(t *testing.T) {
	migrationSQL := readLowerFile(t, internalAnalyticsMigrationPath)
	querySQL := readLowerFile(t, internalAnalyticsQueriesPath)

	if strings.Contains(migrationSQL, "audit_logs") {
		t.Fatal("internal analytics migration must not alter or reference audit_logs")
	}
	if strings.Contains(querySQL, "audit_logs") {
		t.Fatal("internal analytics queries must stay separated from audit_logs")
	}
}

func TestInternalAnalyticsPhase1SqlcQueryContract(t *testing.T) {
	sql := readLowerFile(t, internalAnalyticsQueriesPath)

	required := []string{
		"-- name: createinternalanalyticsevent :one",
		"insert into internal_analytics_events",
		"-- name: getinternalanalyticsevent :one",
		"from internal_analytics_events",
		"-- name: listinternalanalyticseventsforrollup :many",
		"where occurred_at >= sqlc.arg(start_at)",
		"occurred_at < sqlc.arg(end_at)",
		"limit least(greatest(sqlc.arg(limit_count)::int, 1), 10000)",
		"-- name: deleteexpiredinternalanalyticsevents :execrows",
		"delete from internal_analytics_events",
		"retention_expires_at <= sqlc.arg(cutoff_at)",
		"-- name: upsertinternalanalyticsdailyaggregate :one",
		"insert into internal_analytics_daily_aggregates",
		"on conflict (aggregate_date, event_group, event_name, source_surface, role, result)",
		"do update set",
		"-- name: listinternalanalyticsdailyaggregates :many",
		"from internal_analytics_daily_aggregates",
	}

	for _, needle := range required {
		if !strings.Contains(sql, needle) {
			t.Fatalf("internal analytics sqlc query contract missing %q", needle)
		}
	}
}
