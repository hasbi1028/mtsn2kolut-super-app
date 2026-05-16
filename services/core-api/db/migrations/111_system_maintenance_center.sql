-- Maintenance Center: global/module/read-only maintenance windows and audit log.

CREATE TABLE IF NOT EXISTS system_maintenance_windows (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title text NOT NULL,
  message text NOT NULL,
  mode text NOT NULL CHECK (mode IN ('global', 'module', 'read_only')),
  affected_modules text[] NOT NULL DEFAULT ARRAY['global']::text[],
  starts_at timestamptz,
  ends_at timestamptz,
  is_active boolean NOT NULL DEFAULT false,
  allow_admin_bypass boolean NOT NULL DEFAULT true,
  bypass_roles text[] NOT NULL DEFAULT ARRAY['superadmin','admin']::text[],
  severity text NOT NULL DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'critical')),
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  updated_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT system_maintenance_window_time_chk
    CHECK (starts_at IS NULL OR ends_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX IF NOT EXISTS idx_system_maintenance_windows_active
  ON system_maintenance_windows (is_active, starts_at, ends_at);

CREATE INDEX IF NOT EXISTS idx_system_maintenance_windows_created_at
  ON system_maintenance_windows (created_at DESC);

CREATE TABLE IF NOT EXISTS system_maintenance_audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  maintenance_id uuid REFERENCES system_maintenance_windows(id) ON DELETE SET NULL,
  actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  action text NOT NULL,
  reason text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_system_maintenance_audit_logs_created_at
  ON system_maintenance_audit_logs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_system_maintenance_audit_logs_action_created_at
  ON system_maintenance_audit_logs (action, created_at DESC);
