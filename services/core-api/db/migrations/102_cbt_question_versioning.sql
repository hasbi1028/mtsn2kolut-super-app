-- Bank Soal question version lineage.
-- Existing package/session relations already point at exact cbt_questions.id rows;
-- these columns add lineage metadata without rewriting historical references.

ALTER TABLE cbt_questions
  ADD COLUMN IF NOT EXISTS version_group_id UUID REFERENCES cbt_questions(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS version_number INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS source_question_id UUID REFERENCES cbt_questions(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS supersedes_question_id UUID REFERENCES cbt_questions(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS is_latest_version BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS version_note TEXT NOT NULL DEFAULT '';

UPDATE cbt_questions
SET version_group_id = COALESCE(version_group_id, id)
WHERE version_group_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_cbt_questions_version_group
  ON cbt_questions (version_group_id, version_number DESC);

CREATE INDEX IF NOT EXISTS idx_cbt_questions_latest_version
  ON cbt_questions (is_latest_version, workflow_status, status);

CREATE INDEX IF NOT EXISTS idx_cbt_questions_source_question
  ON cbt_questions (source_question_id);

CREATE INDEX IF NOT EXISTS idx_cbt_questions_supersedes_question
  ON cbt_questions (supersedes_question_id);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_cbt_questions_version_number_positive'
      AND conrelid = 'cbt_questions'::regclass
  ) THEN
    ALTER TABLE cbt_questions
      ADD CONSTRAINT chk_cbt_questions_version_number_positive
      CHECK (version_number >= 1);
  END IF;
END $$;
