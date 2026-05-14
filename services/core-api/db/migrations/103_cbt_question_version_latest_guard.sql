-- Guard Bank Soal question version lineage so each group has at most one latest version.
-- Safe after migration 102 backfilled version_group_id for existing questions.

DO $$
DECLARE
  duplicate_count INTEGER;
BEGIN
  SELECT COUNT(*)::int
  INTO duplicate_count
  FROM (
    SELECT version_group_id
    FROM cbt_questions
    WHERE is_latest_version = TRUE
    GROUP BY version_group_id
    HAVING COUNT(*) > 1
  ) dup;

  IF duplicate_count > 0 THEN
    RAISE EXCEPTION 'Cannot create latest-version guard: % version groups have more than one latest question', duplicate_count;
  END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_cbt_questions_one_latest_per_group
  ON cbt_questions (version_group_id)
  WHERE is_latest_version = TRUE;
