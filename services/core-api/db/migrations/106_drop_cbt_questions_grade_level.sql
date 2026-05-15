-- Drop legacy CBT question numeric grade metadata.
-- target_level (VII/VIII/IX) is now the single source of truth.
-- Rollback, if ever needed:
--   ALTER TABLE cbt_questions ADD COLUMN IF NOT EXISTS grade_level SMALLINT;
--   UPDATE cbt_questions
--   SET grade_level = CASE target_level WHEN 'VII' THEN 7 WHEN 'VIII' THEN 8 WHEN 'IX' THEN 9 ELSE NULL END;
--   CREATE INDEX IF NOT EXISTS idx_cbt_questions_grade_level ON cbt_questions (grade_level, subject_id);

ALTER TABLE cbt_questions
  DROP CONSTRAINT IF EXISTS chk_cbt_questions_target_level;

ALTER TABLE cbt_questions
  ADD CONSTRAINT chk_cbt_questions_target_level
  CHECK (
    target_level IS NULL
    OR btrim(target_level) = ''
    OR target_level IN ('VII', 'VIII', 'IX')
  );

DROP INDEX IF EXISTS idx_cbt_questions_grade_level;

ALTER TABLE cbt_questions
  DROP COLUMN IF EXISTS grade_level;
