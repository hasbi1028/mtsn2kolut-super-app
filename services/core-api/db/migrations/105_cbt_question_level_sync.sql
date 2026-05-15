-- Normalize CBT question level metadata.
-- target_level (VII/VIII/IX) is the official source of truth.
-- grade_level remains derived/internal for temporary compatibility.

ALTER TABLE cbt_questions
  ADD COLUMN IF NOT EXISTS target_level TEXT;

UPDATE cbt_questions
SET target_level = CASE grade_level
  WHEN 7 THEN 'VII'
  WHEN 8 THEN 'VIII'
  WHEN 9 THEN 'IX'
  ELSE NULL
END
WHERE NULLIF(btrim(COALESCE(target_level, '')), '') IS NULL
  AND grade_level IN (7, 8, 9);

UPDATE cbt_questions
SET target_level = UPPER(btrim(target_level))
WHERE target_level IS NOT NULL
  AND target_level <> UPPER(btrim(target_level));

UPDATE cbt_questions
SET grade_level = CASE target_level
  WHEN 'VII' THEN 7
  WHEN 'VIII' THEN 8
  WHEN 'IX' THEN 9
  ELSE NULL
END
WHERE target_level IN ('VII', 'VIII', 'IX')
  AND grade_level IS DISTINCT FROM CASE target_level
    WHEN 'VII' THEN 7
    WHEN 'VIII' THEN 8
    WHEN 'IX' THEN 9
  END;

ALTER TABLE cbt_questions
  DROP CONSTRAINT IF EXISTS chk_cbt_questions_target_level;

ALTER TABLE cbt_questions
  ADD CONSTRAINT chk_cbt_questions_target_level
  CHECK (
    target_level IS NULL
    OR btrim(target_level) = ''
    OR target_level IN ('VII', 'VIII', 'IX')
  );
