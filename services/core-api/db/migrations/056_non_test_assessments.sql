-- Migration: 056_non_test_assessments
-- Separate non-test assessment workflow from live CBT question/exam runtime.

CREATE TABLE non_test_assessments (
  id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id            UUID        NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT,
  class_id              UUID        REFERENCES school_classes(id) ON DELETE SET NULL,
  assessment_type       TEXT        NOT NULL CHECK (assessment_type IN ('praktik', 'portofolio', 'proyek', 'penugasan', 'observasi', 'lainnya')),
  title                 TEXT        NOT NULL CHECK (btrim(title) <> ''),
  description           TEXT        NOT NULL DEFAULT '',
  instruction_html      TEXT        NOT NULL DEFAULT '',
  rubric_html           TEXT        NOT NULL DEFAULT '',
  evidence_requirements TEXT        NOT NULL DEFAULT '',
  mode                  TEXT        NOT NULL DEFAULT 'beginner' CHECK (mode IN ('beginner', 'advance')),
  scoring_scale         TEXT        NOT NULL DEFAULT '0_100',
  max_score             NUMERIC(6,2) NOT NULL DEFAULT 100 CHECK (max_score > 0),
  weight                NUMERIC(6,2) NOT NULL DEFAULT 1 CHECK (weight > 0),
  due_at                TIMESTAMPTZ,
  status                TEXT        NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'closed', 'archived')),
  created_by_username   TEXT        NOT NULL DEFAULT '',
  assessor_username     TEXT        NOT NULL DEFAULT '',
  checklist             JSONB       NOT NULL DEFAULT '[]'::jsonb,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE non_test_assessment_submissions (
  id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  assessment_id      UUID        NOT NULL REFERENCES non_test_assessments(id) ON DELETE CASCADE,
  student_id         UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  status             TEXT        NOT NULL DEFAULT 'assigned' CHECK (status IN ('assigned', 'submitted', 'reviewed', 'returned')),
  evidence_url       TEXT        NOT NULL DEFAULT '',
  evidence_note      TEXT        NOT NULL DEFAULT '',
  score              NUMERIC(6,2) CHECK (score IS NULL OR score >= 0),
  feedback           TEXT        NOT NULL DEFAULT '',
  submitted_at       TIMESTAMPTZ,
  graded_at          TIMESTAMPTZ,
  graded_by_username TEXT        NOT NULL DEFAULT '',
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (assessment_id, student_id)
);

CREATE INDEX idx_non_test_assessments_subject_status
  ON non_test_assessments (subject_id, status, created_at DESC);

CREATE INDEX idx_non_test_assessments_class_status
  ON non_test_assessments (class_id, status, due_at);

CREATE INDEX idx_non_test_assessments_type
  ON non_test_assessments (assessment_type, created_at DESC);

CREATE INDEX idx_non_test_submissions_assessment_status
  ON non_test_assessment_submissions (assessment_id, status, student_id);

CREATE INDEX idx_non_test_submissions_student
  ON non_test_assessment_submissions (student_id, updated_at DESC);
