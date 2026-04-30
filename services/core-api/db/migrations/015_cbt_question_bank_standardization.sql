ALTER TABLE cbt_questions
  ADD COLUMN IF NOT EXISTS stem_html TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS stem_latex TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS stimulus_html TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS stimulus_latex TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS explanation_html TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS rubric_html TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS academic_phase TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS grade_level SMALLINT,
  ADD COLUMN IF NOT EXISTS cp_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS tp_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS kd_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS indicator_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS material_topic TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS cognitive_level TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS hots_flag BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS media_asset_ids JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS workflow_status TEXT NOT NULL DEFAULT 'draft',
  ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS author_username TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS reviewer_username TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS approver_username TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS writer_notes TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS review_notes TEXT NOT NULL DEFAULT '';

ALTER TABLE cbt_questions
  DROP CONSTRAINT IF EXISTS chk_cbt_questions_workflow_status;

ALTER TABLE cbt_questions
  ADD CONSTRAINT chk_cbt_questions_workflow_status
  CHECK (workflow_status IN ('draft', 'review', 'approved', 'rejected'));

ALTER TABLE cbt_questions
  DROP CONSTRAINT IF EXISTS chk_cbt_questions_version_positive;

ALTER TABLE cbt_questions
  ADD CONSTRAINT chk_cbt_questions_version_positive
  CHECK (version >= 1);

CREATE INDEX IF NOT EXISTS idx_cbt_questions_workflow_status ON cbt_questions (workflow_status, status);
CREATE INDEX IF NOT EXISTS idx_cbt_questions_grade_level ON cbt_questions (grade_level, subject_id);
CREATE INDEX IF NOT EXISTS idx_cbt_questions_curriculum ON cbt_questions (cp_ref, kd_ref);
CREATE INDEX IF NOT EXISTS idx_cbt_questions_topic_hots ON cbt_questions (material_topic, hots_flag);

CREATE TABLE IF NOT EXISTS cbt_question_assets (
  id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id    UUID        REFERENCES cbt_questions(id) ON DELETE SET NULL,
  original_name  TEXT        NOT NULL,
  stored_name    TEXT        NOT NULL,
  mime_type      TEXT        NOT NULL,
  file_size      BIGINT      NOT NULL,
  storage_path   TEXT        NOT NULL,
  purpose        TEXT        NOT NULL DEFAULT 'general',
  uploaded_by    TEXT        NOT NULL DEFAULT '',
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cbt_question_assets_question_id ON cbt_question_assets (question_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_cbt_question_assets_purpose ON cbt_question_assets (purpose, created_at DESC);
