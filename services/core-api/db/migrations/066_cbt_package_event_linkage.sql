-- Migration 066: Link CBT packages to unified exam events.

ALTER TABLE cbt_packages
  ADD COLUMN IF NOT EXISTS event_id UUID REFERENCES cbt_exam_events(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_cbt_packages_event_subject
  ON cbt_packages (event_id, subject_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_cbt_packages_global_subject
  ON cbt_packages (subject_id, created_at DESC)
  WHERE event_id IS NULL;
