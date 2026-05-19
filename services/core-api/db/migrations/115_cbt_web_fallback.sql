-- CBT browser emergency fallback policy.
-- Default is intentionally OFF per room; native Flutter/desktop clients remain unchanged.

ALTER TABLE cbt_exam_rooms
  ADD COLUMN IF NOT EXISTS allow_web_fallback BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS web_fallback_enabled_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS web_fallback_enabled_by UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS web_fallback_reason TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS web_fallback_disabled_at TIMESTAMPTZ;

ALTER TABLE cbt_exam_participants
  ADD COLUMN IF NOT EXISTS client_type TEXT NOT NULL DEFAULT 'unknown',
  ADD COLUMN IF NOT EXISTS browser_fingerprint_hash TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS client_user_agent_hash TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_cbt_exam_participants_client_type'
  ) THEN
    ALTER TABLE cbt_exam_participants
      ADD CONSTRAINT chk_cbt_exam_participants_client_type
      CHECK (client_type IN ('unknown', 'android', 'windows', 'web_fallback')) NOT VALID;
  END IF;
END $$;

ALTER TABLE cbt_exam_participants
  VALIDATE CONSTRAINT chk_cbt_exam_participants_client_type;

CREATE INDEX IF NOT EXISTS idx_cbt_exam_rooms_web_fallback
  ON cbt_exam_rooms (session_id, allow_web_fallback)
  WHERE allow_web_fallback = TRUE;

CREATE INDEX IF NOT EXISTS idx_cbt_exam_participants_client_type
  ON cbt_exam_participants (session_id, client_type);
