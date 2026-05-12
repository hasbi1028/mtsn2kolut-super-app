-- Milestone 4 KMA 1503: deskripsi capaian per siswa/mapel untuk cetak rapor.
-- Pengaturan rapor sudah tersedia pada migration 089_curriculum_kma_foundation.sql.

CREATE TABLE IF NOT EXISTS grade_student_subject_descriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  assignment_id UUID NOT NULL REFERENCES class_subject_assignments(id) ON DELETE CASCADE,
  student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_grade_student_subject_description UNIQUE (assignment_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_grade_descriptions_student
  ON grade_student_subject_descriptions(student_id);
