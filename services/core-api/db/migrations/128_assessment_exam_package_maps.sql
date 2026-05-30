-- Simple CBT package mapping for the assessment command center.
-- One shared package can be assigned to many rombels for the same subject slot.

CREATE TABLE IF NOT EXISTS assessment_exam_package_maps (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_id    UUID        NOT NULL REFERENCES assessment_exams(id) ON DELETE CASCADE,
  class_id   UUID        NOT NULL REFERENCES school_classes(id) ON DELETE CASCADE,
  subject_id UUID        NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT,
  package_id UUID        NOT NULL REFERENCES cbt_packages(id) ON DELETE RESTRICT,
  slot_label TEXT        NOT NULL DEFAULT '',
  notes      TEXT        NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (exam_id, class_id, subject_id)
);

CREATE INDEX IF NOT EXISTS idx_assessment_exam_package_maps_exam
  ON assessment_exam_package_maps (exam_id, subject_id, class_id);

CREATE INDEX IF NOT EXISTS idx_assessment_exam_package_maps_package
  ON assessment_exam_package_maps (package_id);
