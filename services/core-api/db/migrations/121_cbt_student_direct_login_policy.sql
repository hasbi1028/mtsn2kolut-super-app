ALTER TABLE cbt_exam_sessions
  ADD COLUMN IF NOT EXISTS access_mode TEXT NOT NULL DEFAULT 'secure_exam',
  ADD COLUMN IF NOT EXISTS student_portal_direct_login_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS require_room_token_for_web BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS nisn_direct_login_enabled BOOLEAN NOT NULL DEFAULT FALSE;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_cbt_exam_sessions_access_mode'
      AND conrelid = 'cbt_exam_sessions'::regclass
  ) THEN
    ALTER TABLE cbt_exam_sessions
      ADD CONSTRAINT chk_cbt_exam_sessions_access_mode
      CHECK (access_mode IN ('secure_exam', 'web_fallback', 'simulation')) NOT VALID;
  END IF;
END $$;

ALTER TABLE cbt_exam_sessions
  VALIDATE CONSTRAINT chk_cbt_exam_sessions_access_mode;

CREATE INDEX IF NOT EXISTS idx_cbt_exam_sessions_access_policy
  ON cbt_exam_sessions (status, access_mode, student_portal_direct_login_enabled, require_room_token_for_web, nisn_direct_login_enabled);
