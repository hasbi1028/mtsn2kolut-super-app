-- Migration 083: Internal analytics phase 1 schema.
-- Contract-only schema draft; runtime ingestion is intentionally added in a later phase.

CREATE TABLE IF NOT EXISTS internal_analytics_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_name TEXT NOT NULL,
  event_group TEXT NOT NULL,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  source_surface TEXT NOT NULL,
  actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  actor_role TEXT,
  permission_code TEXT,
  route_group TEXT,
  module TEXT,
  result TEXT,
  status_code_class TEXT,
  duration_bucket TEXT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  retention_expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_internal_analytics_events_group CHECK (event_group IN ('public', 'auth', 'dashboard', 'bank_soal', 'asesmen', 'pusaka', 'users_rbac', 'security')),
  CONSTRAINT chk_internal_analytics_events_name CHECK (event_name ~ '^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$'),
  CONSTRAINT chk_internal_analytics_events_source CHECK (btrim(source_surface) <> ''),
  CONSTRAINT chk_internal_analytics_events_metadata_object CHECK (jsonb_typeof(metadata) = 'object'),
  CONSTRAINT chk_internal_analytics_events_retention CHECK (retention_expires_at > occurred_at)
);

CREATE INDEX IF NOT EXISTS idx_internal_analytics_events_occurred_at
  ON internal_analytics_events (occurred_at);

CREATE INDEX IF NOT EXISTS idx_internal_analytics_events_retention_expires_at
  ON internal_analytics_events (retention_expires_at);

CREATE INDEX IF NOT EXISTS idx_internal_analytics_events_group_occurred_at
  ON internal_analytics_events (event_group, occurred_at);

CREATE INDEX IF NOT EXISTS idx_internal_analytics_events_name_occurred_at
  ON internal_analytics_events (event_name, occurred_at);

CREATE TABLE IF NOT EXISTS internal_analytics_daily_aggregates (
  aggregate_date DATE NOT NULL,
  event_group TEXT NOT NULL,
  event_name TEXT NOT NULL,
  source_surface TEXT NOT NULL DEFAULT '',
  role TEXT NOT NULL DEFAULT '',
  result TEXT NOT NULL DEFAULT '',
  count BIGINT NOT NULL DEFAULT 0,
  total_duration_ms BIGINT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_internal_analytics_daily_aggregates_group CHECK (event_group IN ('public', 'auth', 'dashboard', 'bank_soal', 'asesmen', 'pusaka', 'users_rbac', 'security')),
  CONSTRAINT chk_internal_analytics_daily_aggregates_name CHECK (event_name ~ '^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$'),
  CONSTRAINT chk_internal_analytics_daily_aggregates_count CHECK (count >= 0),
  CONSTRAINT chk_internal_analytics_daily_aggregates_duration CHECK (total_duration_ms IS NULL OR total_duration_ms >= 0),
  CONSTRAINT chk_internal_analytics_daily_aggregates_metadata_object CHECK (jsonb_typeof(metadata) = 'object'),
  CONSTRAINT uq_internal_analytics_daily_aggregates_key UNIQUE (aggregate_date, event_group, event_name, source_surface, role, result)
);

CREATE INDEX IF NOT EXISTS idx_internal_analytics_daily_aggregates_date
  ON internal_analytics_daily_aggregates (aggregate_date);

CREATE INDEX IF NOT EXISTS idx_internal_analytics_daily_aggregates_group_date
  ON internal_analytics_daily_aggregates (event_group, aggregate_date);
