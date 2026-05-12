CREATE TABLE IF NOT EXISTS class_subject_allocation_overrides (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  assignment_id UUID NOT NULL REFERENCES class_subject_assignments(id) ON DELETE CASCADE,
  curriculum_allocation_id UUID REFERENCES curriculum_subject_allocations(id) ON DELETE SET NULL,
  intra_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  koku_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  additional_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  total_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  is_customized BOOLEAN NOT NULL DEFAULT FALSE,
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_assignment_allocation_override UNIQUE (assignment_id),
  CONSTRAINT chk_assignment_allocation_hours_nonnegative CHECK (
    intra_weekly_hours >= 0
    AND koku_weekly_hours >= 0
    AND additional_weekly_hours >= 0
    AND total_weekly_hours >= 0
  )
);

CREATE INDEX IF NOT EXISTS idx_class_subject_allocation_overrides_allocation
  ON class_subject_allocation_overrides(curriculum_allocation_id);
