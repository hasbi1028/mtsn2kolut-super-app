CREATE TYPE cbt_event_member_role AS ENUM (
  'panitia',
  'pembuat_soal',
  'reviewer',
  'proktor',
  'pengawas',
  'korektor'
);

CREATE TABLE cbt_event_members (
  id          UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id    UUID                  NOT NULL REFERENCES cbt_exam_events(id) ON DELETE CASCADE,
  user_id     UUID                  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  employee_id UUID                  REFERENCES employees(id) ON DELETE SET NULL,
  subject_id  UUID                  REFERENCES subjects(id) ON DELETE CASCADE,
  role        cbt_event_member_role NOT NULL,
  created_at  TIMESTAMPTZ           NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ           NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_cbt_event_members_employee_subject_scope
    CHECK (role IN ('pembuat_soal', 'reviewer', 'korektor') OR subject_id IS NULL)
);

CREATE UNIQUE INDEX uq_cbt_event_members_assignment
  ON cbt_event_members (event_id, user_id, role, COALESCE(subject_id, '00000000-0000-0000-0000-000000000000'::uuid));

CREATE INDEX idx_cbt_event_members_event_role ON cbt_event_members (event_id, role, subject_id);
CREATE INDEX idx_cbt_event_members_user ON cbt_event_members (user_id, event_id, role);

ALTER TABLE cbt_questions
  ADD COLUMN IF NOT EXISTS event_id UUID REFERENCES cbt_exam_events(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_cbt_questions_event_subject ON cbt_questions (event_id, subject_id, workflow_status, status);
