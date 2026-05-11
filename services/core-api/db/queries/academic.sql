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
SELECT
    id,
    code,
    name,
    category,
    is_assessment_subject,
    is_report_subject,
    is_schedule_activity,
    default_weekly_hours,
    display_order,
    is_active,
    created_at,
    updated_at
FROM subjects
ORDER BY display_order ASC, name ASC;

-- name: CreateSubject :one
INSERT INTO subjects (
    id,
    code,
    name,
    category,
    is_assessment_subject,
    is_report_subject,
    is_schedule_activity,
    default_weekly_hours,
    display_order,
    is_active
)
VALUES (
    gen_random_uuid(),
    sqlc.arg(code),
    sqlc.arg(name),
    sqlc.arg(category),
    sqlc.arg(is_assessment_subject),
    sqlc.arg(is_report_subject),
    sqlc.arg(is_schedule_activity),
    sqlc.arg(default_weekly_hours),
    sqlc.arg(display_order),
    sqlc.arg(is_active)
)
RETURNING *;

-- name: GetSubject :one
SELECT
    id,
    code,
    name,
    category,
    is_assessment_subject,
    is_report_subject,
    is_schedule_activity,
    default_weekly_hours,
    display_order,
    is_active,
    created_at,
    updated_at
FROM subjects
WHERE id = $1;

-- name: CountSubjectCodeConflicts :one
SELECT COUNT(*)::int
FROM subjects
WHERE id <> sqlc.arg(id)
  AND LOWER(code) = LOWER(sqlc.arg(code));

-- name: UpdateSubject :one
UPDATE subjects
SET code = sqlc.arg(code),
    name = sqlc.arg(name),
    category = sqlc.arg(category),
    is_assessment_subject = sqlc.arg(is_assessment_subject),
    is_report_subject = sqlc.arg(is_report_subject),
    is_schedule_activity = sqlc.arg(is_schedule_activity),
    default_weekly_hours = sqlc.arg(default_weekly_hours),
    display_order = sqlc.arg(display_order),
    is_active = sqlc.arg(is_active),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteSubject :exec
DELETE FROM subjects WHERE id = $1;

-- name: GetAcademicStats :one
SELECT 
    (SELECT COUNT(*) FROM students WHERE is_active = TRUE)::int AS total_students,
    (SELECT COUNT(*) FROM school_classes WHERE is_active = TRUE)::int AS total_classes,
    (SELECT COUNT(*) FROM subjects WHERE is_active = TRUE)::int AS total_subjects,
    (SELECT COUNT(*) FROM academic_years)::int AS total_years;

-- name: GetAcademicDashboardSummary :one
WITH active_year AS (
    SELECT id, name
    FROM academic_years
    WHERE is_active = TRUE
    ORDER BY start_date DESC, name DESC
    LIMIT 1
),
active_classes AS (
    SELECT c.*
    FROM school_classes c
    WHERE c.is_active = TRUE
      AND (
          NOT EXISTS (SELECT 1 FROM active_year)
          OR c.academic_year_id = (SELECT id FROM active_year)
      )
),
active_students AS (
    SELECT s.*
    FROM students s
    WHERE s.is_active = TRUE
),
active_assignments AS (
    SELECT csa.*
    FROM class_subject_assignments csa
    JOIN active_classes c ON c.id = csa.class_id
),
dashboard_slots AS (
    SELECT
        ts.id,
        ts.day_of_week,
        ts.start_time,
        ts.end_time,
        csa.class_id,
        csa.teacher_employee_id,
        LOWER(TRIM(ts.room_label)) AS room_key
    FROM timetable_slots ts
    JOIN active_assignments csa ON csa.id = ts.assignment_id
)
SELECT
    COALESCE((SELECT name FROM active_year), '')::text AS active_academic_year,
    CASE
        WHEN EXTRACT(MONTH FROM CURRENT_DATE)::int BETWEEN 7 AND 12 THEN 'Ganjil'
        ELSE 'Genap'
    END::text AS active_semester,
    (SELECT COUNT(*) FROM active_classes)::int AS total_classes,
    (SELECT COUNT(*) FROM active_students)::int AS total_active_students,
    (
        SELECT COUNT(*)
        FROM active_students s
        WHERE s.class_id IS NULL
           OR NOT EXISTS (SELECT 1 FROM active_classes c WHERE c.id = s.class_id)
    )::int AS students_without_class,
    (
        SELECT COUNT(*)
        FROM active_classes c
        WHERE NOT EXISTS (
            SELECT 1
            FROM class_homeroom_assignments cha
            WHERE cha.class_id = c.id
              AND cha.is_active = TRUE
        )
    )::int AS classes_without_homeroom,
    (
        SELECT COUNT(*)
        FROM active_assignments csa
        LEFT JOIN employees e ON e.id = csa.teacher_employee_id
        WHERE e.id IS NULL OR e.is_active = FALSE
    )::int AS subject_assignments_missing_teacher,
    (
        SELECT COUNT(DISTINCT a.id)
        FROM dashboard_slots a
        JOIN dashboard_slots b ON a.id < b.id
         AND a.day_of_week = b.day_of_week
         AND a.start_time < b.end_time
         AND a.end_time > b.start_time
         AND (
            a.class_id = b.class_id
            OR a.teacher_employee_id = b.teacher_employee_id
            OR (a.room_key <> '' AND a.room_key = b.room_key)
         )
    )::int AS timetable_conflicts,
    (
        SELECT COUNT(*)
        FROM active_students s
        WHERE NOT EXISTS (
            SELECT 1
            FROM users u
            WHERE u.student_id = s.id
              AND u.deleted_at IS NULL
              AND u.is_active = TRUE
        )
    )::int AS student_accounts_missing,
    (
        SELECT COUNT(DISTINCT p.id)
        FROM parent_students ps
        JOIN active_students s ON s.id = ps.student_id
        JOIN parents p ON p.id = ps.parent_id
        WHERE NOT EXISTS (
            SELECT 1
            FROM users u
            WHERE u.parent_id = p.id
              AND u.deleted_at IS NULL
              AND u.is_active = TRUE
        )
    )::int AS parent_accounts_missing;

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

-- name: GetActiveAcademicYear :one
SELECT id, name, start_date, end_date, is_active, created_at, updated_at
FROM academic_years
WHERE is_active = TRUE
ORDER BY start_date DESC, name DESC
LIMIT 1;

-- name: ListSubjectAssignmentMatrixClasses :many
SELECT id, code, name, level
FROM school_classes
WHERE academic_year_id = $1
  AND is_active = TRUE
ORDER BY level ASC, name ASC;

-- name: ListSubjectAssignmentMatrixSubjects :many
SELECT
    id,
    code,
    name,
    category,
    is_assessment_subject,
    is_report_subject,
    is_schedule_activity,
    default_weekly_hours,
    display_order
FROM subjects
WHERE is_active = TRUE
ORDER BY display_order ASC, name ASC;

-- name: ListSubjectAssignmentMatrixTeachers :many
SELECT
    id,
    COALESCE(nip, '')::text AS nip,
    nama,
    unit_kerja
FROM employees
WHERE is_active = TRUE
ORDER BY nama ASC;

-- name: ListSubjectAssignmentMatrixCells :many
SELECT
    c.id AS class_id,
    s.id AS subject_id,
    csa.id AS assignment_id,
    csa.teacher_employee_id,
    COALESCE(e.nama, '') AS teacher_name,
    CASE
        WHEN csa.id IS NULL THEN 'missing_assignment'
        WHEN e.id IS NULL OR e.is_active = FALSE THEN 'missing_teacher'
        ELSE 'complete'
    END::text AS status
FROM school_classes c
CROSS JOIN subjects s
LEFT JOIN class_subject_assignments csa ON csa.class_id = c.id AND csa.subject_id = s.id
LEFT JOIN employees e ON e.id = csa.teacher_employee_id
WHERE c.academic_year_id = $1
  AND c.is_active = TRUE
  AND s.is_active = TRUE
ORDER BY s.display_order ASC, s.name ASC, c.level ASC, c.name ASC;

-- name: GetSubjectAssignmentMatrixCell :one
SELECT
    c.id AS class_id,
    s.id AS subject_id,
    csa.id AS assignment_id,
    csa.teacher_employee_id,
    COALESCE(e.nama, '') AS teacher_name,
    CASE
        WHEN csa.id IS NULL THEN 'missing_assignment'
        WHEN e.id IS NULL OR e.is_active = FALSE THEN 'missing_teacher'
        ELSE 'complete'
    END::text AS status
FROM school_classes c
CROSS JOIN subjects s
LEFT JOIN class_subject_assignments csa ON csa.class_id = c.id AND csa.subject_id = s.id
LEFT JOIN employees e ON e.id = csa.teacher_employee_id
JOIN academic_years ay ON ay.id = c.academic_year_id AND ay.is_active = TRUE
WHERE c.id = sqlc.arg(class_id)
  AND c.is_active = TRUE
  AND s.id = sqlc.arg(subject_id)
  AND s.is_active = TRUE;

-- name: GetSubjectAssignmentByClassSubject :one
SELECT id, class_id, subject_id, teacher_employee_id, created_at, updated_at
FROM class_subject_assignments
WHERE class_id = sqlc.arg(class_id)
  AND subject_id = sqlc.arg(subject_id);

-- name: UpsertSubjectAssignmentMatrixCell :one
WITH validated AS (
    SELECT
        sqlc.arg(class_id)::uuid AS class_id,
        sqlc.arg(subject_id)::uuid AS subject_id,
        sqlc.arg(teacher_employee_id)::uuid AS teacher_employee_id
    WHERE EXISTS (
        SELECT 1
        FROM school_classes c
        JOIN academic_years ay ON ay.id = c.academic_year_id AND ay.is_active = TRUE
        WHERE c.id = sqlc.arg(class_id)::uuid
          AND c.is_active = TRUE
    )
      AND EXISTS (
        SELECT 1
        FROM subjects s
        WHERE s.id = sqlc.arg(subject_id)::uuid
          AND s.is_active = TRUE
      )
      AND EXISTS (
        SELECT 1
        FROM employees e
        WHERE e.id = sqlc.arg(teacher_employee_id)::uuid
          AND e.is_active = TRUE
      )
)
INSERT INTO class_subject_assignments (id, class_id, subject_id, teacher_employee_id)
SELECT gen_random_uuid(), class_id, subject_id, teacher_employee_id
FROM validated
ON CONFLICT (class_id, subject_id)
DO UPDATE SET teacher_employee_id = EXCLUDED.teacher_employee_id,
              updated_at = NOW()
RETURNING *;
