CREATE TABLE IF NOT EXISTS cbt_event_subject_targets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id UUID NOT NULL REFERENCES cbt_exam_events(id) ON DELETE CASCADE,
  subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT,
  target_questions INTEGER NOT NULL CHECK (target_questions > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (event_id, subject_id)
);

CREATE INDEX IF NOT EXISTS idx_cbt_event_subject_targets_event
  ON cbt_event_subject_targets (event_id, subject_id);

CREATE TABLE IF NOT EXISTS cbt_question_audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id UUID NOT NULL REFERENCES cbt_questions(id) ON DELETE CASCADE,
  actor_username TEXT NOT NULL DEFAULT '',
  action TEXT NOT NULL,
  note TEXT NOT NULL DEFAULT '',
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_cbt_question_audit_action CHECK (action IN (
    'create',
    'update',
    'import',
    'submit_review',
    'approve',
    'reject',
    'publish',
    'duplicate',
    'revision',
    'delete'
  ))
);

CREATE INDEX IF NOT EXISTS idx_cbt_question_audit_logs_question_created
  ON cbt_question_audit_logs (question_id, created_at DESC);
