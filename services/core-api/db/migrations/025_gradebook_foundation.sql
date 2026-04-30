CREATE TABLE IF NOT EXISTS grade_components (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  assignment_id UUID NOT NULL REFERENCES class_subject_assignments(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT 'assignment',
  weight DOUBLE PRECISION NOT NULL DEFAULT 1,
  max_score DOUBLE PRECISION NOT NULL DEFAULT 100,
  is_published BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_grade_components_category
    CHECK (category IN ('assignment', 'quiz', 'midterm', 'final', 'project', 'practice', 'attitude', 'attendance', 'other')),
  CONSTRAINT chk_grade_components_weight_nonnegative
    CHECK (weight >= 0),
  CONSTRAINT chk_grade_components_max_score_positive
    CHECK (max_score > 0)
);

CREATE INDEX IF NOT EXISTS idx_grade_components_assignment
  ON grade_components (assignment_id, created_at DESC);

CREATE TABLE IF NOT EXISTS grade_entries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  component_id UUID NOT NULL REFERENCES grade_components(id) ON DELETE CASCADE,
  student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  score DOUBLE PRECISION,
  notes TEXT NOT NULL DEFAULT '',
  graded_by TEXT NOT NULL DEFAULT '',
  graded_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_grade_entries_component_student UNIQUE (component_id, student_id),
  CONSTRAINT chk_grade_entries_score_nonnegative
    CHECK (score IS NULL OR score >= 0)
);

CREATE INDEX IF NOT EXISTS idx_grade_entries_component_student
  ON grade_entries (component_id, student_id);
