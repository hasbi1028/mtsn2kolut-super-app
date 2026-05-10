ALTER TABLE cbt_exam_participants
  ADD COLUMN IF NOT EXISTS violation_count integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS risk_score integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS risk_level text NOT NULL DEFAULT 'normal',
  ADD COLUMN IF NOT EXISTS locked_at timestamptz,
  ADD COLUMN IF NOT EXISTS locked_reason text;

ALTER TABLE cbt_exam_participants
  ADD CONSTRAINT cbt_exam_participants_risk_level_check
  CHECK (risk_level IN ('normal', 'warning', 'high', 'locked')) NOT VALID;

ALTER TABLE cbt_exam_participants
  VALIDATE CONSTRAINT cbt_exam_participants_risk_level_check;

CREATE INDEX IF NOT EXISTS idx_cbt_exam_participants_risk
  ON cbt_exam_participants (session_id, risk_level, locked_at);
