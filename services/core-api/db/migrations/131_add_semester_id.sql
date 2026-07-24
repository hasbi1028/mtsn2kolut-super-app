-- Migration: 131_add_semester_id
-- Tambah kolom semester_id ke tabel data agar bisa difilter per semester aktif.

-- 1. Tambah semester_id ke tabel-tabel data
ALTER TABLE class_subject_assignments ADD COLUMN semester_id UUID REFERENCES semesters(id) ON DELETE SET NULL;
ALTER TABLE class_homeroom_assignments ADD COLUMN semester_id UUID REFERENCES semesters(id) ON DELETE SET NULL;
ALTER TABLE timetable_slots ADD COLUMN semester_id UUID REFERENCES semesters(id) ON DELETE SET NULL;
ALTER TABLE grade_components ADD COLUMN semester_id UUID REFERENCES semesters(id) ON DELETE SET NULL;
ALTER TABLE grade_entries ADD COLUMN semester_id UUID REFERENCES semesters(id) ON DELETE SET NULL;
ALTER TABLE grade_assignment_finalizations ADD COLUMN semester_id UUID REFERENCES semesters(id) ON DELETE SET NULL;
ALTER TABLE class_journal ADD COLUMN semester_id UUID REFERENCES semesters(id) ON DELETE SET NULL;
ALTER TABLE non_test_assessment_submissions ADD COLUMN semester_id UUID REFERENCES semesters(id) ON DELETE SET NULL;
ALTER TABLE cbt_exam_events ADD COLUMN semester_id UUID REFERENCES semesters(id) ON DELETE SET NULL;
ALTER TABLE class_curriculum_assignments ADD COLUMN semester_id UUID REFERENCES semesters(id) ON DELETE SET NULL;

-- 2. Set semester_id ke semester aktif untuk data yang sudah ada (jika ada)
-- Ambil ID semester aktif
DO $$
DECLARE
    active_sem_id UUID;
BEGIN
    SELECT id INTO active_sem_id FROM semesters WHERE is_active = TRUE LIMIT 1;
    
    IF active_sem_id IS NOT NULL THEN
        UPDATE class_subject_assignments SET semester_id = active_sem_id WHERE semester_id IS NULL;
        UPDATE class_homeroom_assignments SET semester_id = active_sem_id WHERE semester_id IS NULL;
        UPDATE timetable_slots SET semester_id = active_sem_id WHERE semester_id IS NULL;
        UPDATE grade_components SET semester_id = active_sem_id WHERE semester_id IS NULL;
        UPDATE grade_entries SET semester_id = active_sem_id WHERE semester_id IS NULL;
        UPDATE grade_assignment_finalizations SET semester_id = active_sem_id WHERE semester_id IS NULL;
        UPDATE class_journal SET semester_id = active_sem_id WHERE semester_id IS NULL;
        UPDATE non_test_assessment_submissions SET semester_id = active_sem_id WHERE semester_id IS NULL;
        UPDATE cbt_exam_events SET semester_id = active_sem_id WHERE semester_id IS NULL;
        UPDATE class_curriculum_assignments SET semester_id = active_sem_id WHERE semester_id IS NULL;
    END IF;
END $$;

-- 3. Index untuk performa filter per semester
CREATE INDEX idx_class_subject_assignments_semester ON class_subject_assignments(semester_id);
CREATE INDEX idx_timetable_slots_semester ON timetable_slots(semester_id);
CREATE INDEX idx_grade_components_semester ON grade_components(semester_id);
CREATE INDEX idx_grade_entries_semester ON grade_entries(semester_id);
CREATE INDEX idx_grade_assignment_finalizations_semester ON grade_assignment_finalizations(semester_id);
CREATE INDEX idx_class_journal_semester ON class_journal(semester_id);
CREATE INDEX idx_cbt_exam_events_semester ON cbt_exam_events(semester_id);
CREATE INDEX idx_class_curriculum_assignments_semester ON class_curriculum_assignments(semester_id);
