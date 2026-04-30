CREATE TABLE IF NOT EXISTS grade_assignment_finalizations (
  assignment_id UUID PRIMARY KEY REFERENCES class_subject_assignments(id) ON DELETE CASCADE,
  finalized_by TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  finalized_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_grade_assignment_finalizations_finalized_at
  ON grade_assignment_finalizations (finalized_at DESC);
