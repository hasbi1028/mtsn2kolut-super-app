-- Bank Soal 3-status workflow simplification.
--
-- Keep publication state in cbt_questions.status (draft/published/archived),
-- while workflow_status becomes the user-facing work status:
-- konsep -> diperiksa -> siap_pakai.

ALTER TABLE cbt_questions
  DROP CONSTRAINT IF EXISTS chk_cbt_questions_workflow_status;

UPDATE cbt_questions
SET workflow_status = CASE
  WHEN workflow_status IN ('draft', 'revision', 'revision_needed', 'rejected') THEN 'konsep'
  WHEN workflow_status IN ('review', 'submitted', 'reviewed') THEN 'diperiksa'
  WHEN workflow_status IN ('approved', 'published') THEN 'siap_pakai'
  WHEN workflow_status = 'archived' THEN 'konsep'
  WHEN workflow_status IN ('konsep', 'diperiksa', 'siap_pakai') THEN workflow_status
  ELSE 'konsep'
END;

ALTER TABLE cbt_questions
  ALTER COLUMN workflow_status SET DEFAULT 'konsep';

ALTER TABLE cbt_questions
  ADD CONSTRAINT chk_cbt_questions_workflow_status
  CHECK (workflow_status IN ('konsep', 'diperiksa', 'siap_pakai'));

DROP INDEX IF EXISTS idx_cbt_questions_workflow_status;
CREATE INDEX IF NOT EXISTS idx_cbt_questions_workflow_status
  ON cbt_questions (workflow_status, status);
