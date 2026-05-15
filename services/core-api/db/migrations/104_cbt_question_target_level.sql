ALTER TABLE cbt_questions
  ADD COLUMN IF NOT EXISTS target_level TEXT;

UPDATE cbt_questions
SET target_level = ''
WHERE target_level IS NULL;

ALTER TABLE cbt_questions
  DROP CONSTRAINT IF EXISTS chk_cbt_questions_target_level;

ALTER TABLE cbt_questions
  ADD CONSTRAINT chk_cbt_questions_target_level
  CHECK (target_level IS NULL OR target_level = '' OR target_level IN ('VII', 'VIII', 'IX'));

CREATE INDEX IF NOT EXISTS idx_cbt_questions_target_level
  ON cbt_questions (target_level, subject_id, status, workflow_status);
