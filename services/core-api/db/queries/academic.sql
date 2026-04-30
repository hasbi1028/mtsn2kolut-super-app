-- name: ListAcademicYears :many
SELECT id, name, start_date, end_date, is_active, created_at, updated_at
FROM academic_years
ORDER BY start_date DESC, name DESC;

-- name: CreateAcademicYear :one
INSERT INTO academic_years (id, name, start_date, end_date, is_active)
VALUES (gen_random_uuid(), $1, $2, $3, $4)
RETURNING *;

-- name: DeleteAcademicYear :exec
DELETE FROM academic_years WHERE id = $1;

-- name: ListSchoolClasses :many
SELECT c.id, c.code, c.name, c.level, c.is_active, c.created_at, c.updated_at,
       c.academic_year_id, a.name AS academic_year_name
FROM school_classes c
JOIN academic_years a ON a.id = c.academic_year_id
ORDER BY a.start_date DESC, c.level ASC, c.name ASC;

-- name: CreateSchoolClass :one
INSERT INTO school_classes (id, academic_year_id, code, name, level, is_active)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteSchoolClass :exec
DELETE FROM school_classes WHERE id = $1;

-- name: ListSubjects :many
SELECT id, code, name, is_active, created_at, updated_at
FROM subjects
ORDER BY name ASC;

-- name: CreateSubject :one
INSERT INTO subjects (id, code, name, is_active)
VALUES (gen_random_uuid(), $1, $2, $3)
RETURNING *;

-- name: DeleteSubject :exec
DELETE FROM subjects WHERE id = $1;

-- name: GetAcademicStats :one
SELECT 
    (SELECT COUNT(*) FROM students WHERE is_active = TRUE)::int AS total_students,
    (SELECT COUNT(*) FROM school_classes WHERE is_active = TRUE)::int AS total_classes,
    (SELECT COUNT(*) FROM subjects WHERE is_active = TRUE)::int AS total_subjects,
    (SELECT COUNT(*) FROM academic_years)::int AS total_years;

-- name: ListClassSubjectAssignments :many
SELECT a.id, a.class_id, c.name AS class_name, c.code AS class_code,
       a.subject_id, s.name AS subject_name, s.code AS subject_code,
       a.teacher_employee_id, e.nama AS teacher_name,
       a.created_at, a.updated_at
FROM class_subject_assignments a
JOIN school_classes c ON c.id = a.class_id
JOIN subjects s ON s.id = a.subject_id
JOIN employees e ON e.id = a.teacher_employee_id
ORDER BY c.name ASC, s.name ASC;

-- name: CreateClassSubjectAssignment :one
INSERT INTO class_subject_assignments (id, class_id, subject_id, teacher_employee_id)
VALUES (gen_random_uuid(), $1, $2, $3)
RETURNING *;

-- name: DeleteClassSubjectAssignment :exec
DELETE FROM class_subject_assignments WHERE id = $1;

-- name: GetClassSubjectAssignment :one
SELECT a.id, a.class_id, c.name AS class_name, c.code AS class_code,
       a.subject_id, s.name AS subject_name, s.code AS subject_code,
       a.teacher_employee_id, e.nama AS teacher_name,
       a.created_at, a.updated_at
FROM class_subject_assignments a
JOIN school_classes c ON c.id = a.class_id
JOIN subjects s ON s.id = a.subject_id
JOIN employees e ON e.id = a.teacher_employee_id
WHERE a.id = $1;
