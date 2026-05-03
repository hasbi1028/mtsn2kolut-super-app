-- Migration: 057_non_test_grade_sync
-- Link non-test assessments to official gradebook components for repeatable sync.

ALTER TABLE non_test_assessments
  ADD COLUMN grade_component_id UUID REFERENCES grade_components(id) ON DELETE SET NULL,
  ADD COLUMN grade_synced_at TIMESTAMPTZ,
  ADD COLUMN grade_synced_by TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_non_test_assessments_grade_component
  ON non_test_assessments (grade_component_id)
  WHERE grade_component_id IS NOT NULL;
