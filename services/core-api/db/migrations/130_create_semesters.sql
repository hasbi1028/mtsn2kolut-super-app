-- Migration: 130_create_semesters
-- Tambah tabel semesters, bersihkan data akademik lama, seed fresh.

-- 1. Buat tabel semesters
CREATE TABLE semesters (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    academic_year_id UUID        NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    name            TEXT         NOT NULL,              -- 'Ganjil' / 'Genap'
    label           TEXT         NOT NULL,              -- 'Semester Ganjil 2025/2026'
    start_date      DATE         NOT NULL,
    end_date        DATE         NOT NULL,
    is_active       BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (academic_year_id, name)
);

CREATE UNIQUE INDEX uq_semesters_active_true
    ON semesters (is_active)
    WHERE is_active = TRUE;

-- 2. Hapus unique index active lama (pindah ke semesters)
DROP INDEX IF EXISTS uq_academic_years_active_true;

-- 3. Hapus data lama (berurutan sesuai FK)
DELETE FROM assessment_participants;
DELETE FROM cbt_exam_sessions;
DELETE FROM assessment_exam_package_maps;
DELETE FROM assessment_exam_seats;
DELETE FROM non_test_assessment_submissions;
DELETE FROM non_test_assessment_responses;
DELETE FROM grade_entries;
DELETE FROM grade_components;
DELETE FROM grade_assignment_finalizations;
DELETE FROM grade_student_subject_descriptions;

DELETE FROM lesson_period_templates;
DELETE FROM report_settings;
DELETE FROM class_curriculum_assignments;
DELETE FROM curriculum_profiles;
DELETE FROM class_subject_allocation_overrides;
DELETE FROM class_subject_assignments;
DELETE FROM class_homeroom_assignments;
DELETE FROM timetable_slots;
DELETE FROM students;
DELETE FROM parent_students;
DELETE FROM student_id_card_scans;
DELETE FROM student_id_cards;
DELETE FROM student_certificates;

DELETE FROM school_classes;
DELETE FROM rombel_students;
DELETE FROM rombel_teachers;
DELETE FROM rombel_supervisors;
DELETE FROM rombel_members;

DELETE FROM cbt_exam_events;

-- 4. Hapus semua tahun akademik dan seed ulang
DELETE FROM academic_years;

INSERT INTO academic_years (id, name, start_date, end_date, is_active)
VALUES (gen_random_uuid(), '2025/2026', DATE '2025-07-01', DATE '2026-06-30', TRUE);

-- 5. Rombel default
WITH target_year AS (
    SELECT id FROM academic_years WHERE name = '2025/2026' LIMIT 1
)
INSERT INTO school_classes (academic_year_id, code, name, level, is_active)
SELECT target_year.id, seed.code, seed.name, seed.level, TRUE
FROM target_year
CROSS JOIN (
    VALUES
        ('VII.A', 'VII.A', 'VII'),
        ('VII.B', 'VII.B', 'VII'),
        ('VII.C', 'VII.C', 'VII'),
        ('VII.D', 'VII.D', 'VII'),
        ('VIII.A', 'VIII.A', 'VIII'),
        ('VIII.B', 'VIII.B', 'VIII'),
        ('VIII.C', 'VIII.C', 'VIII'),
        ('IX.A',  'IX.A',  'IX'),
        ('IX.B',  'IX.B',  'IX'),
        ('IX.C',  'IX.C',  'IX')
) AS seed(code, name, level);

-- 6. Seed 2 semester
WITH target_year AS (
    SELECT id FROM academic_years WHERE name = '2025/2026' LIMIT 1
)
INSERT INTO semesters (academic_year_id, name, label, start_date, end_date, is_active)
SELECT
    target_year.id,
    seed.name,
    seed.label,
    seed.start_date,
    seed.end_date,
    seed.is_active
FROM target_year
CROSS JOIN (
    VALUES
        ('Ganjil', 'Semester Ganjil 2025/2026', DATE '2025-07-01', DATE '2025-12-31', TRUE),
        ('Genap',  'Semester Genap 2025/2026',  DATE '2026-01-01', DATE '2026-06-30', FALSE)
) AS seed(name, label, start_date, end_date, is_active);
