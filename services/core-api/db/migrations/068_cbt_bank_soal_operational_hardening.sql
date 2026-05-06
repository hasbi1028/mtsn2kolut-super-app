-- Bank Soal/Asesmen operational hardening.
-- Keep question audit history available even when an unused draft question is deleted,
-- and add lookup indexes used by Bank Soal usage counters on large datasets.

ALTER TABLE cbt_question_audit_logs
  ALTER COLUMN question_id DROP NOT NULL;

ALTER TABLE cbt_question_audit_logs
  DROP CONSTRAINT IF EXISTS cbt_question_audit_logs_question_id_fkey;

ALTER TABLE cbt_question_audit_logs
  ADD CONSTRAINT cbt_question_audit_logs_question_id_fkey
  FOREIGN KEY (question_id) REFERENCES cbt_questions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_cbt_package_questions_question_id
  ON cbt_package_questions (question_id);

CREATE INDEX IF NOT EXISTS idx_cbt_student_answers_question_id
  ON cbt_student_answers (question_id);
