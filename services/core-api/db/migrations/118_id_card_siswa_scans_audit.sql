-- Wave 1: ID Card scan/audit integration tables.

CREATE TABLE IF NOT EXISTS student_activity_attendance_scans (
  id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  card_id        UUID        NOT NULL REFERENCES student_id_cards(id) ON DELETE RESTRICT,
  student_id     UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  activity_code  TEXT        NOT NULL DEFAULT '',
  scan_type      TEXT        NOT NULL DEFAULT 'checkin',
  scanned_by_user_id UUID    REFERENCES users(id) ON DELETE SET NULL,
  scanned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  metadata       JSONB       NOT NULL DEFAULT '{}'::jsonb,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_student_activity_attendance_scan_type CHECK (scan_type IN ('checkin','checkout','present','late','out'))
);

CREATE INDEX IF NOT EXISTS idx_student_activity_attendance_scans_student ON student_activity_attendance_scans(student_id, scanned_at DESC);
CREATE INDEX IF NOT EXISTS idx_student_activity_attendance_scans_activity ON student_activity_attendance_scans(activity_code, scanned_at DESC);

CREATE TABLE IF NOT EXISTS student_id_card_audit_logs (
  id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  card_id        UUID        REFERENCES student_id_cards(id) ON DELETE SET NULL,
  student_id     UUID        REFERENCES students(id) ON DELETE SET NULL,
  actor_user_id  UUID        REFERENCES users(id) ON DELETE SET NULL,
  action         TEXT        NOT NULL,
  target         TEXT        NOT NULL DEFAULT 'student_id_card',
  metadata       JSONB       NOT NULL DEFAULT '{}'::jsonb,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_student_id_card_audit_logs_card ON student_id_card_audit_logs(card_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_student_id_card_audit_logs_student ON student_id_card_audit_logs(student_id, created_at DESC);
