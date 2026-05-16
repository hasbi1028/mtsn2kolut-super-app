-- Sprint 2 Bank Soal workflow actions.
-- Backward-compatible: keep legacy "review" status as a submitted/review alias
-- and add workflow event auditing without rewriting existing question rows.

ALTER TABLE cbt_questions
  DROP CONSTRAINT IF EXISTS chk_cbt_questions_workflow_status;

ALTER TABLE cbt_questions
  ADD CONSTRAINT chk_cbt_questions_workflow_status
  CHECK (
    workflow_status IN (
      'draft',
      'review',
      'submitted',
      'revision_needed',
      'reviewed',
      'approved',
      'published',
      'rejected',
      'archived'
    )
  );

CREATE TABLE IF NOT EXISTS bank_soal_question_workflow_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id UUID NOT NULL REFERENCES cbt_questions(id) ON DELETE CASCADE,
  actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  actor_username TEXT NOT NULL DEFAULT '',
  from_status TEXT NOT NULL DEFAULT '',
  to_status TEXT NOT NULL,
  action TEXT NOT NULL,
  note TEXT NOT NULL DEFAULT '',
  metadata JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bank_soal_workflow_events_question
  ON bank_soal_question_workflow_events(question_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_bank_soal_workflow_events_actor
  ON bank_soal_question_workflow_events(actor_username, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_bank_soal_workflow_events_action
  ON bank_soal_question_workflow_events(action, created_at DESC);
