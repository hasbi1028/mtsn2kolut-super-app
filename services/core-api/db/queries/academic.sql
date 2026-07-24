-- name: ListAcademicYears :many
SELECT id, name, start_date, end_date, is_active, created_at, updated_at
FROM academic_years
ORDER BY start_date DESC, name DESC;

-- name: GetAcademicYearByID :one
SELECT id, name, start_date, end_date, is_active, created_at, updated_at
FROM academic_years
WHERE id = $1;

-- name: CountAcademicYearNameConflicts :one
SELECT COUNT(*)::int
FROM academic_years
WHERE LOWER(name) = LOWER(sqlc.arg(name));

-- name: CreateAcademicYear :one
INSERT INTO academic_years (id, name, start_date, end_date, is_active)
VALUES (gen_random_uuid(), $1, $2, $3, $4)
RETURNING *;

-- name: DeactivateAcademicYears :exec
UPDATE academic_years
SET is_active = FALSE,
    updated_at = NOW()
WHERE is_active = TRUE;

-- name: ActivateAcademicYear :one
UPDATE academic_years
SET is_active = TRUE,
    updated_at = NOW()
WHERE id = $1
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
    counts_for_ranking,
    is_local_content,
    is_choice_subject,
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
    counts_for_ranking,
    is_local_content,
    is_choice_subject,
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
    sqlc.arg(counts_for_ranking),
    sqlc.arg(is_local_content),
    sqlc.arg(is_choice_subject),
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
    counts_for_ranking,
    is_local_content,
    is_choice_subject,
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
    counts_for_ranking = sqlc.arg(counts_for_ranking),
    is_local_content = sqlc.arg(is_local_content),
    is_choice_subject = sqlc.arg(is_choice_subject),
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
active_class_curriculum AS (
    SELECT DISTINCT ON (cca.class_id)
        cca.class_id,
        cca.curriculum_profile_id
    FROM class_curriculum_assignments cca
    JOIN active_classes c ON c.id = cca.class_id
    WHERE cca.is_active = TRUE
    ORDER BY cca.class_id, cca.updated_at DESC, cca.created_at DESC
),
required_allocations AS (
    SELECT
        c.id AS class_id,
        c.level,
        alloc.id AS allocation_id,
        alloc.subject_id,
        alloc.subject_group,
        alloc.total_weekly_hours,
        alloc.counts_for_schedule,
        alloc.counts_for_report,
        alloc.counts_for_ranking,
        alloc.is_required
    FROM active_classes c
    JOIN active_class_curriculum acc ON acc.class_id = c.id
    JOIN curriculum_subject_allocations alloc
      ON alloc.curriculum_profile_id = acc.curriculum_profile_id
     AND alloc.level = c.level
     AND alloc.is_required = TRUE
),
assignment_effective_hours AS (
    SELECT
        csa.id AS assignment_id,
        csa.class_id,
        csa.subject_id,
        csa.teacher_employee_id,
        COALESCE(ov.total_weekly_hours, alloc.total_weekly_hours, 0)::numeric(6,2) AS total_weekly_hours,
        COALESCE(ov.additional_weekly_hours, 0)::numeric(6,2) AS additional_weekly_hours
    FROM active_assignments csa
    JOIN active_classes c ON c.id = csa.class_id
    LEFT JOIN active_class_curriculum acc ON acc.class_id = c.id
    LEFT JOIN curriculum_subject_allocations alloc
      ON alloc.curriculum_profile_id = acc.curriculum_profile_id
     AND alloc.level = c.level
     AND alloc.subject_id = csa.subject_id
    LEFT JOIN class_subject_allocation_overrides ov ON ov.assignment_id = csa.id
),
class_curriculum_hours AS (
    SELECT
        c.id AS class_id,
        COALESCE((
            SELECT SUM(alloc.total_weekly_hours)
            FROM curriculum_subject_allocations alloc
            JOIN active_class_curriculum acc ON acc.curriculum_profile_id = alloc.curriculum_profile_id
            WHERE acc.class_id = c.id
              AND alloc.level = c.level
              AND alloc.is_required = TRUE
        ), 0)::numeric(6,2)
        + COALESCE((
            SELECT SUM(aeh.additional_weekly_hours)
            FROM assignment_effective_hours aeh
            WHERE aeh.class_id = c.id
        ), 0)::numeric(6,2) AS effective_weekly_hours
    FROM active_classes c
),
dashboard_slots AS (
    SELECT
        ts.id,
        ts.assignment_id,
        ts.day_of_week,
        ts.start_time,
        ts.end_time,
        ts.lesson_hours,
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
    )::int AS parent_accounts_missing,
    (
        SELECT COUNT(*)
        FROM active_classes c
        WHERE NOT EXISTS (SELECT 1 FROM active_class_curriculum acc WHERE acc.class_id = c.id)
    )::int AS classes_without_curriculum_profile,
    (
        SELECT COUNT(*)
        FROM class_curriculum_hours cch
        WHERE cch.effective_weekly_hours < 42
    )::int AS classes_weekly_hours_under_42,
    (
        SELECT COUNT(*)
        FROM class_curriculum_hours cch
        WHERE cch.effective_weekly_hours > 48
    )::int AS classes_weekly_hours_over_48,
    (
        SELECT COUNT(*)
        FROM required_allocations alloc
        LEFT JOIN active_assignments csa ON csa.class_id = alloc.class_id AND csa.subject_id = alloc.subject_id
        LEFT JOIN employees e ON e.id = csa.teacher_employee_id AND e.is_active = TRUE
        WHERE alloc.counts_for_schedule = TRUE
          AND e.id IS NULL
    )::int AS required_subjects_missing_teacher,
    (
        SELECT COUNT(*)
        FROM (
            SELECT e.id, COALESCE(SUM(aeh.total_weekly_hours), 0)::numeric(6,2) AS total_weekly_hours
            FROM employees e
            JOIN active_assignments csa ON csa.teacher_employee_id = e.id
            LEFT JOIN assignment_effective_hours aeh ON aeh.assignment_id = csa.id
            WHERE e.is_active = TRUE
            GROUP BY e.id
        ) teacher_load
        WHERE teacher_load.total_weekly_hours < 24
    )::int AS teachers_under_24_hours,
    (
        SELECT COUNT(*)
        FROM assignment_effective_hours aeh
        LEFT JOIN (
            SELECT assignment_id, COALESCE(SUM(lesson_hours), 0)::numeric(6,2) AS scheduled_weekly_hours
            FROM dashboard_slots
            GROUP BY assignment_id
        ) scheduled ON scheduled.assignment_id = aeh.assignment_id
        WHERE ABS(COALESCE(scheduled.scheduled_weekly_hours, 0) - aeh.total_weekly_hours) > 0.01
    )::int AS timetable_hours_mismatch,
    (
        SELECT COUNT(*)
        FROM subjects s
        WHERE s.is_active = TRUE
          AND s.counts_for_ranking = FALSE
          AND s.is_report_subject = TRUE
    )::int AS non_ranking_subjects_in_ranking;


-- name: GetAcademicReadinessSummary :one
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
    JOIN active_year ay ON ay.id = c.academic_year_id
    WHERE c.is_active = TRUE
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
report_assignments AS (
    SELECT csa.*
    FROM active_assignments csa
    JOIN subjects s ON s.id = csa.subject_id
    WHERE s.is_active = TRUE
      AND s.is_report_subject = TRUE
),
active_class_curriculum AS (
    SELECT DISTINCT ON (cca.class_id)
        cca.class_id,
        cca.curriculum_profile_id
    FROM class_curriculum_assignments cca
    JOIN active_classes c ON c.id = cca.class_id
    WHERE cca.is_active = TRUE
    ORDER BY cca.class_id, cca.updated_at DESC, cca.created_at DESC
),
required_report_allocations AS (
    SELECT c.id AS class_id, alloc.subject_id
    FROM active_classes c
    JOIN active_class_curriculum acc ON acc.class_id = c.id
    JOIN curriculum_subject_allocations alloc
      ON alloc.curriculum_profile_id = acc.curriculum_profile_id
     AND alloc.level = c.level
    WHERE alloc.is_required = TRUE
      AND alloc.counts_for_report = TRUE
),
slots AS (
    SELECT
        ts.id,
        ts.assignment_id,
        ts.day_of_week,
        ts.start_time,
        ts.end_time,
        csa.class_id,
        csa.teacher_employee_id,
        LOWER(TRIM(ts.room_label)) AS room_key
    FROM timetable_slots ts
    JOIN active_assignments csa ON csa.id = ts.assignment_id
),
teacher_hours AS (
    SELECT
        e.id AS teacher_employee_id,
        COALESCE(SUM(COALESCE(ov.total_weekly_hours, alloc.total_weekly_hours, s.default_weekly_hours::numeric, 0)), 0)::numeric(6,2) AS total_weekly_hours
    FROM employees e
    JOIN active_assignments csa ON csa.teacher_employee_id = e.id
    JOIN active_classes c ON c.id = csa.class_id
    JOIN subjects s ON s.id = csa.subject_id
    LEFT JOIN active_class_curriculum acc ON acc.class_id = c.id
    LEFT JOIN curriculum_subject_allocations alloc
      ON alloc.curriculum_profile_id = acc.curriculum_profile_id
     AND alloc.level = c.level
     AND alloc.subject_id = csa.subject_id
    LEFT JOIN class_subject_allocation_overrides ov ON ov.assignment_id = csa.id
    WHERE e.is_active = TRUE
      AND s.is_active = TRUE
    GROUP BY e.id
),
component_rollup AS (
    SELECT csa.id AS assignment_id,
           COUNT(gc.id)::int AS component_count,
           COUNT(gc.id) FILTER (WHERE gc.is_published = TRUE)::int AS published_component_count
    FROM report_assignments csa
    LEFT JOIN grade_components gc ON gc.assignment_id = csa.id
    GROUP BY csa.id
),
student_component_rollup AS (
    SELECT csa.id AS assignment_id,
           st.id AS student_id,
           COALESCE(cr.component_count, 0)::int AS component_count,
           COUNT(ge.score)::int AS filled_count
    FROM report_assignments csa
    JOIN active_students st ON st.class_id = csa.class_id
    LEFT JOIN component_rollup cr ON cr.assignment_id = csa.id
    LEFT JOIN grade_components gc ON gc.assignment_id = csa.id
    LEFT JOIN grade_entries ge ON ge.component_id = gc.id AND ge.student_id = st.id
    GROUP BY csa.id, st.id, cr.component_count
),
report_description_scope AS (
    SELECT csa.id AS assignment_id, st.id AS student_id
    FROM report_assignments csa
    JOIN active_students st ON st.class_id = csa.class_id
)
SELECT
    COALESCE((SELECT name FROM active_year), '')::text AS active_academic_year,
    CASE WHEN EXTRACT(MONTH FROM CURRENT_DATE)::int BETWEEN 7 AND 12 THEN 'Ganjil' ELSE 'Genap' END::text AS active_semester,
    (SELECT COUNT(*) FROM active_classes)::int AS total_classes,
    (SELECT COUNT(*) FROM active_students)::int AS total_active_students,
    (SELECT COUNT(*) FROM active_classes c WHERE NOT EXISTS (SELECT 1 FROM active_students s WHERE s.class_id = c.id))::int AS classes_without_students,
    (SELECT COUNT(*) FROM active_classes c WHERE NOT EXISTS (SELECT 1 FROM class_homeroom_assignments cha WHERE cha.class_id = c.id AND cha.is_active = TRUE))::int AS classes_without_homeroom,
    (SELECT COUNT(*) FROM active_classes c WHERE NOT EXISTS (SELECT 1 FROM active_class_curriculum acc WHERE acc.class_id = c.id))::int AS classes_without_curriculum_profile,
    (SELECT COUNT(*) FROM required_report_allocations alloc LEFT JOIN report_assignments csa ON csa.class_id = alloc.class_id AND csa.subject_id = alloc.subject_id LEFT JOIN employees e ON e.id = csa.teacher_employee_id AND e.is_active = TRUE WHERE e.id IS NULL)::int AS report_subjects_missing_teacher,
    (SELECT COUNT(DISTINCT a.id) FROM slots a JOIN slots b ON a.id < b.id AND a.day_of_week = b.day_of_week AND a.start_time < b.end_time AND a.end_time > b.start_time AND (a.class_id = b.class_id OR a.teacher_employee_id = b.teacher_employee_id OR (a.room_key <> '' AND a.room_key = b.room_key)))::int AS timetable_conflicts,
    (SELECT COUNT(*) FROM teacher_hours WHERE total_weekly_hours < 24)::int AS teachers_under_24_hours,
    (SELECT COUNT(*) FROM teacher_hours WHERE total_weekly_hours > 40)::int AS teachers_over_40_hours,
    (SELECT COUNT(*) FROM report_assignments csa WHERE NOT EXISTS (SELECT 1 FROM component_rollup cr WHERE cr.assignment_id = csa.id AND cr.component_count > 0))::int AS report_assignments_without_components,
    (SELECT COUNT(*) FROM student_component_rollup WHERE component_count = 0 OR filled_count < component_count)::int AS students_with_incomplete_grades,
    (SELECT COUNT(*) FROM report_assignments csa WHERE NOT EXISTS (SELECT 1 FROM grade_assignment_finalizations gaf WHERE gaf.assignment_id = csa.id))::int AS report_assignments_not_finalized,
    (SELECT COUNT(*) FROM report_description_scope rds LEFT JOIN grade_student_subject_descriptions gd ON gd.assignment_id = rds.assignment_id AND gd.student_id = rds.student_id WHERE gd.id IS NULL OR NULLIF(BTRIM(gd.description), '') IS NULL)::int AS report_descriptions_missing,
    (SELECT COUNT(*) FROM report_settings rs JOIN active_year ay ON ay.id = rs.academic_year_id)::int AS report_settings_count;

-- name: ListTeacherWorkload :many
WITH active_classes AS (
    SELECT id, code, name, level
    FROM school_classes
    WHERE academic_year_id = $1
      AND is_active = TRUE
),
active_class_curriculum AS (
    SELECT DISTINCT ON (cca.class_id)
        cca.class_id,
        cca.curriculum_profile_id
    FROM class_curriculum_assignments cca
    JOIN active_classes c ON c.id = cca.class_id
    WHERE cca.is_active = TRUE
    ORDER BY cca.class_id, cca.updated_at DESC, cca.created_at DESC
),
scheduled_by_assignment AS (
    SELECT
        ts.assignment_id,
        COALESCE(SUM(ts.lesson_hours), 0)::numeric(6,2) AS scheduled_weekly_hours
    FROM timetable_slots ts
    JOIN class_subject_assignments csa ON csa.id = ts.assignment_id
    JOIN active_classes c ON c.id = csa.class_id
    GROUP BY ts.assignment_id
),
assignment_hours AS (
    SELECT
        csa.id AS assignment_id,
        e.id AS teacher_employee_id,
        COALESCE(e.nip, '')::text AS teacher_nip,
        e.nama AS teacher_name,
        COALESCE(e.unit_kerja, '')::text AS unit_kerja,
        csa.class_id,
        c.code AS class_code,
        c.name AS class_name,
        c.level AS class_level,
        csa.subject_id,
        s.code AS subject_code,
        s.name AS subject_name,
        COALESCE(alloc.subject_group, CASE WHEN s.is_schedule_activity THEN 'kegiatan' ELSE s.category END)::text AS subject_group,
        COALESCE(ov.intra_weekly_hours, alloc.intra_weekly_hours, CASE WHEN s.is_schedule_activity THEN 0 ELSE s.default_weekly_hours::numeric END, 0)::numeric(6,2) AS intra_weekly_hours,
        COALESCE(ov.koku_weekly_hours, alloc.koku_weekly_hours, CASE WHEN s.is_schedule_activity THEN s.default_weekly_hours::numeric ELSE 0 END, 0)::numeric(6,2) AS koku_weekly_hours,
        COALESCE(ov.additional_weekly_hours, 0)::numeric(6,2) AS additional_weekly_hours,
        COALESCE(ov.total_weekly_hours, alloc.total_weekly_hours, s.default_weekly_hours::numeric, 0)::numeric(6,2) AS total_weekly_hours,
        COALESCE(sba.scheduled_weekly_hours, 0)::numeric(6,2) AS scheduled_weekly_hours
    FROM class_subject_assignments csa
    JOIN active_classes c ON c.id = csa.class_id
    JOIN subjects s ON s.id = csa.subject_id
    JOIN employees e ON e.id = csa.teacher_employee_id
    LEFT JOIN active_class_curriculum acc ON acc.class_id = c.id
    LEFT JOIN curriculum_subject_allocations alloc
      ON alloc.curriculum_profile_id = acc.curriculum_profile_id
     AND alloc.level = c.level
     AND alloc.subject_id = csa.subject_id
    LEFT JOIN class_subject_allocation_overrides ov ON ov.assignment_id = csa.id
    LEFT JOIN scheduled_by_assignment sba ON sba.assignment_id = csa.id
    WHERE s.is_active = TRUE
      AND e.is_active = TRUE
)
SELECT
    teacher_employee_id,
    teacher_nip,
    teacher_name,
    unit_kerja,
    COUNT(DISTINCT class_id)::int AS total_classes,
    COUNT(DISTINCT subject_id)::int AS total_subjects,
    COUNT(*)::int AS total_assignments,
    COALESCE(SUM(intra_weekly_hours), 0)::float8 AS intra_weekly_hours,
    COALESCE(SUM(koku_weekly_hours), 0)::float8 AS koku_weekly_hours,
    COALESCE(SUM(additional_weekly_hours), 0)::float8 AS additional_weekly_hours,
    0::float8 AS coordination_equivalent_hours,
    COALESCE(SUM(total_weekly_hours), 0)::float8 AS total_weekly_hours,
    COALESCE(SUM(scheduled_weekly_hours), 0)::float8 AS scheduled_weekly_hours,
    CASE
        WHEN COALESCE(SUM(total_weekly_hours), 0) < 24 THEN 'kurang'
        WHEN COALESCE(SUM(total_weekly_hours), 0) > 40 THEN 'lebih'
        ELSE 'cukup'
    END::text AS workload_status
FROM assignment_hours
GROUP BY teacher_employee_id, teacher_nip, teacher_name, unit_kerja
ORDER BY teacher_name ASC;

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

-- name: ListYearRolloverStudents :many
SELECT
    s.id,
    s.nis,
    s.nisn,
    s.nama,
    s.class_id,
    c.code AS class_code,
    c.name AS class_name,
    c.level AS class_level
FROM students s
JOIN school_classes c ON c.id = s.class_id
WHERE c.academic_year_id = $1
  AND s.is_active = TRUE
  AND s.status = 'active'
ORDER BY c.level ASC, c.name ASC, s.nama ASC;

-- name: ListYearRolloverHomeroomAssignments :many
SELECT
    cha.class_id,
    COUNT(*)::int AS total
FROM class_homeroom_assignments cha
JOIN school_classes c ON c.id = cha.class_id
WHERE c.academic_year_id = $1
  AND c.is_active = TRUE
  AND cha.is_active = TRUE
GROUP BY cha.class_id;

-- name: ListYearRolloverHomeroomAssignmentDetails :many
SELECT
    cha.id,
    cha.class_id,
    cha.employee_id,
    cha.academic_year_id,
    cha.start_date,
    cha.end_date,
    cha.is_active,
    cha.notes
FROM class_homeroom_assignments cha
JOIN school_classes c ON c.id = cha.class_id
WHERE c.academic_year_id = $1
  AND c.is_active = TRUE
  AND cha.is_active = TRUE
ORDER BY c.level ASC, c.name ASC, cha.start_date DESC;

-- name: CountActiveHomeroomAssignmentByClass :one
SELECT COUNT(*)::int
FROM class_homeroom_assignments
WHERE class_id = $1
  AND is_active = TRUE;

-- name: PromoteYearRolloverStudent :execrows
UPDATE students
SET class_id = sqlc.arg(target_class_id),
    updated_at = NOW()
WHERE id = sqlc.arg(student_id)
  AND class_id = sqlc.arg(source_class_id)
  AND is_active = TRUE
  AND status = 'active';

-- name: ListAcademicImportStudents :many
SELECT
    id,
    nis,
    nisn,
    nama,
    class_id,
    is_active,
    status
FROM students
ORDER BY nama ASC;

-- name: ListAcademicImportTeachers :many
SELECT
    id,
    COALESCE(nip, '')::text AS nip,
    nama,
    is_active
FROM employees
WHERE is_active = TRUE
ORDER BY nama ASC;

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
    counts_for_ranking,
    is_local_content,
    is_choice_subject,
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
WITH active_classes AS (
    SELECT id, code, name, level
    FROM school_classes
    WHERE academic_year_id = $1
      AND is_active = TRUE
),
active_class_curriculum AS (
    SELECT DISTINCT ON (cca.class_id)
        cca.class_id,
        cca.curriculum_profile_id
    FROM class_curriculum_assignments cca
    JOIN active_classes c ON c.id = cca.class_id
    WHERE cca.is_active = TRUE
    ORDER BY cca.class_id, cca.updated_at DESC, cca.created_at DESC
)
SELECT
    c.id AS class_id,
    s.id AS subject_id,
    csa.id AS assignment_id,
    csa.teacher_employee_id,
    COALESCE(e.nama, '') AS teacher_name,
    CASE
        WHEN alloc.id IS NULL THEN 'missing_curriculum'
        WHEN csa.id IS NULL THEN 'missing_assignment'
        WHEN e.id IS NULL OR e.is_active = FALSE THEN 'missing_teacher'
        ELSE 'complete'
    END::text AS status,
    alloc.id AS curriculum_allocation_id,
    COALESCE(ov.intra_weekly_hours, alloc.intra_weekly_hours, 0)::numeric(6,2) AS intra_weekly_hours,
    COALESCE(ov.koku_weekly_hours, alloc.koku_weekly_hours, 0)::numeric(6,2) AS koku_weekly_hours,
    COALESCE(ov.additional_weekly_hours, 0)::numeric(6,2) AS additional_weekly_hours,
    COALESCE(ov.total_weekly_hours, alloc.total_weekly_hours, 0)::numeric(6,2) AS total_weekly_hours,
    COALESCE(ov.is_customized, FALSE)::boolean AS is_customized,
    CASE
        WHEN alloc.id IS NULL THEN 'perlu_kurikulum'
        WHEN COALESCE(ov.total_weekly_hours, alloc.total_weekly_hours, 0) = alloc.total_weekly_hours THEN 'sesuai'
        WHEN COALESCE(ov.total_weekly_hours, alloc.total_weekly_hours, 0) < alloc.total_weekly_hours THEN 'kurang'
        ELSE 'lebih'
    END::text AS compliance_status
FROM active_classes c
CROSS JOIN subjects s
LEFT JOIN active_class_curriculum acc ON acc.class_id = c.id
LEFT JOIN curriculum_subject_allocations alloc
  ON alloc.curriculum_profile_id = acc.curriculum_profile_id
 AND alloc.level = c.level
 AND alloc.subject_id = s.id
LEFT JOIN class_subject_assignments csa ON csa.class_id = c.id AND csa.subject_id = s.id
LEFT JOIN employees e ON e.id = csa.teacher_employee_id
LEFT JOIN class_subject_allocation_overrides ov ON ov.assignment_id = csa.id
WHERE s.is_active = TRUE
ORDER BY s.display_order ASC, s.name ASC, c.level ASC, c.name ASC;

-- name: GetSubjectAssignmentMatrixCell :one
WITH active_class AS (
    SELECT c.id, c.code, c.name, c.level
    FROM school_classes c
    JOIN academic_years ay ON ay.id = c.academic_year_id AND ay.is_active = TRUE
    WHERE c.id = sqlc.arg(class_id)
      AND c.is_active = TRUE
),
active_class_curriculum AS (
    SELECT DISTINCT ON (cca.class_id)
        cca.class_id,
        cca.curriculum_profile_id
    FROM class_curriculum_assignments cca
    JOIN active_class c ON c.id = cca.class_id
    WHERE cca.is_active = TRUE
    ORDER BY cca.class_id, cca.updated_at DESC, cca.created_at DESC
)
SELECT
    c.id AS class_id,
    s.id AS subject_id,
    csa.id AS assignment_id,
    csa.teacher_employee_id,
    COALESCE(e.nama, '') AS teacher_name,
    CASE
        WHEN alloc.id IS NULL THEN 'missing_curriculum'
        WHEN csa.id IS NULL THEN 'missing_assignment'
        WHEN e.id IS NULL OR e.is_active = FALSE THEN 'missing_teacher'
        ELSE 'complete'
    END::text AS status,
    alloc.id AS curriculum_allocation_id,
    COALESCE(ov.intra_weekly_hours, alloc.intra_weekly_hours, 0)::numeric(6,2) AS intra_weekly_hours,
    COALESCE(ov.koku_weekly_hours, alloc.koku_weekly_hours, 0)::numeric(6,2) AS koku_weekly_hours,
    COALESCE(ov.additional_weekly_hours, 0)::numeric(6,2) AS additional_weekly_hours,
    COALESCE(ov.total_weekly_hours, alloc.total_weekly_hours, 0)::numeric(6,2) AS total_weekly_hours,
    COALESCE(ov.is_customized, FALSE)::boolean AS is_customized,
    CASE
        WHEN alloc.id IS NULL THEN 'perlu_kurikulum'
        WHEN COALESCE(ov.total_weekly_hours, alloc.total_weekly_hours, 0) = alloc.total_weekly_hours THEN 'sesuai'
        WHEN COALESCE(ov.total_weekly_hours, alloc.total_weekly_hours, 0) < alloc.total_weekly_hours THEN 'kurang'
        ELSE 'lebih'
    END::text AS compliance_status
FROM active_class c
CROSS JOIN subjects s
LEFT JOIN active_class_curriculum acc ON acc.class_id = c.id
LEFT JOIN curriculum_subject_allocations alloc
  ON alloc.curriculum_profile_id = acc.curriculum_profile_id
 AND alloc.level = c.level
 AND alloc.subject_id = s.id
LEFT JOIN class_subject_assignments csa ON csa.class_id = c.id AND csa.subject_id = s.id
LEFT JOIN employees e ON e.id = csa.teacher_employee_id
LEFT JOIN class_subject_allocation_overrides ov ON ov.assignment_id = csa.id
WHERE s.id = sqlc.arg(subject_id)
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

-- name: ListCurriculumProfiles :many
SELECT
    id,
    code,
    name,
    regulation_reference,
    education_level,
    effective_academic_year_id,
    status,
    notes,
    created_at,
    updated_at
FROM curriculum_profiles
ORDER BY status = 'active' DESC, created_at DESC, name ASC;

-- name: GetActiveCurriculumProfile :one
SELECT
    id,
    code,
    name,
    regulation_reference,
    education_level,
    effective_academic_year_id,
    status,
    notes,
    created_at,
    updated_at
FROM curriculum_profiles
WHERE status = 'active'
ORDER BY created_at DESC, name ASC
LIMIT 1;

-- name: ListCurriculumSubjectAllocations :many
SELECT
    csa.id,
    csa.curriculum_profile_id,
    cp.code AS curriculum_profile_code,
    cp.name AS curriculum_profile_name,
    csa.subject_id,
    s.code AS subject_code,
    s.name AS subject_name,
    csa.level,
    csa.subject_group,
    csa.intra_annual_hours,
    csa.koku_annual_hours,
    csa.total_annual_hours,
    csa.intra_weekly_hours,
    csa.koku_weekly_hours,
    csa.total_weekly_hours,
    csa.lesson_minutes,
    csa.display_order,
    csa.counts_for_schedule,
    csa.counts_for_report,
    csa.counts_for_assessment,
    csa.counts_for_ranking,
    csa.is_required,
    csa.notes,
    csa.created_at,
    csa.updated_at
FROM curriculum_subject_allocations csa
JOIN curriculum_profiles cp ON cp.id = csa.curriculum_profile_id
JOIN subjects s ON s.id = csa.subject_id
WHERE csa.curriculum_profile_id = sqlc.arg(curriculum_profile_id)
  AND (sqlc.narg(level)::text IS NULL OR csa.level = sqlc.narg(level)::text)
ORDER BY csa.level ASC, csa.display_order ASC, s.name ASC;

-- name: GetCurriculumSummaryByLevel :many
SELECT
    csa.curriculum_profile_id,
    csa.level,
    COUNT(*)::int AS subject_count,
    COALESCE(SUM(csa.intra_annual_hours), 0)::int AS intra_annual_hours,
    COALESCE(SUM(csa.koku_annual_hours), 0)::int AS koku_annual_hours,
    COALESCE(SUM(csa.total_annual_hours), 0)::int AS total_annual_hours,
    COALESCE(SUM(csa.intra_weekly_hours), 0)::numeric(6,2) AS intra_weekly_hours,
    COALESCE(SUM(csa.koku_weekly_hours), 0)::numeric(6,2) AS koku_weekly_hours,
    COALESCE(SUM(csa.total_weekly_hours), 0)::numeric(6,2) AS total_weekly_hours,
    CASE
        WHEN COALESCE(SUM(csa.total_weekly_hours), 0) = 42 THEN 'sesuai'
        ELSE 'perlu_ditinjau'
    END::text AS compliance_status
FROM curriculum_subject_allocations csa
WHERE csa.curriculum_profile_id = sqlc.arg(curriculum_profile_id)
  AND csa.is_required = TRUE
GROUP BY csa.curriculum_profile_id, csa.level
ORDER BY csa.level ASC;

-- name: ListClassCurriculumAssignments :many
SELECT
    cca.id,
    cca.class_id,
    sc.code AS class_code,
    sc.name AS class_name,
    sc.level AS class_level,
    cca.curriculum_profile_id,
    cp.code AS curriculum_profile_code,
    cp.name AS curriculum_profile_name,
    cca.is_active,
    cca.notes,
    cca.created_at,
    cca.updated_at
FROM class_curriculum_assignments cca
JOIN school_classes sc ON sc.id = cca.class_id
JOIN curriculum_profiles cp ON cp.id = cca.curriculum_profile_id
WHERE (sqlc.narg(curriculum_profile_id)::uuid IS NULL OR cca.curriculum_profile_id = sqlc.narg(curriculum_profile_id)::uuid)
ORDER BY sc.level ASC, sc.name ASC;

-- name: GetCurriculumAllocationForClassSubject :one
SELECT
    alloc.id,
    alloc.curriculum_profile_id,
    alloc.subject_id,
    alloc.level,
    alloc.subject_group,
    alloc.intra_weekly_hours,
    alloc.koku_weekly_hours,
    alloc.total_weekly_hours,
    alloc.counts_for_schedule,
    alloc.is_required
FROM school_classes c
JOIN class_curriculum_assignments cca ON cca.class_id = c.id AND cca.is_active = TRUE
JOIN curriculum_subject_allocations alloc
  ON alloc.curriculum_profile_id = cca.curriculum_profile_id
 AND alloc.level = c.level
 AND alloc.subject_id = sqlc.arg(subject_id)
WHERE c.id = sqlc.arg(class_id)
  AND c.is_active = TRUE
ORDER BY cca.updated_at DESC, cca.created_at DESC
LIMIT 1;

-- name: SumClassAdditionalWeeklyHoursExceptAssignment :one
SELECT COALESCE(SUM(ov.additional_weekly_hours), 0)::numeric(6,2)
FROM class_subject_allocation_overrides ov
JOIN class_subject_assignments csa ON csa.id = ov.assignment_id
WHERE csa.class_id = sqlc.arg(class_id)
  AND csa.id <> sqlc.arg(exclude_assignment_id);

-- name: UpsertClassSubjectAllocationOverride :one
INSERT INTO class_subject_allocation_overrides (
    id,
    assignment_id,
    curriculum_allocation_id,
    intra_weekly_hours,
    koku_weekly_hours,
    additional_weekly_hours,
    total_weekly_hours,
    is_customized,
    notes
)
VALUES (
    gen_random_uuid(),
    sqlc.arg(assignment_id),
    sqlc.narg(curriculum_allocation_id),
    sqlc.arg(intra_weekly_hours),
    sqlc.arg(koku_weekly_hours),
    sqlc.arg(additional_weekly_hours),
    sqlc.arg(total_weekly_hours),
    sqlc.arg(is_customized),
    sqlc.arg(notes)
)
ON CONFLICT (assignment_id)
DO UPDATE SET curriculum_allocation_id = EXCLUDED.curriculum_allocation_id,
              intra_weekly_hours = EXCLUDED.intra_weekly_hours,
              koku_weekly_hours = EXCLUDED.koku_weekly_hours,
              additional_weekly_hours = EXCLUDED.additional_weekly_hours,
              total_weekly_hours = EXCLUDED.total_weekly_hours,
              is_customized = EXCLUDED.is_customized,
              notes = EXCLUDED.notes,
              updated_at = NOW()
RETURNING *;

-- name: ListSemesters :many
SELECT s.id, s.academic_year_id, ay.name AS academic_year_name,
       s.name, s.label, s.start_date, s.end_date, s.is_active,
       s.created_at, s.updated_at
FROM semesters s
JOIN academic_years ay ON ay.id = s.academic_year_id
ORDER BY s.start_date DESC, s.name DESC;

-- name: GetSemester :one
SELECT s.id, s.academic_year_id, ay.name AS academic_year_name,
       s.name, s.label, s.start_date, s.end_date, s.is_active,
       s.created_at, s.updated_at
FROM semesters s
JOIN academic_years ay ON ay.id = s.academic_year_id
WHERE s.id = $1;

-- name: GetActiveSemester :one
SELECT s.id, s.academic_year_id, ay.name AS academic_year_name,
       s.name, s.label, s.start_date, s.end_date, s.is_active,
       s.created_at, s.updated_at
FROM semesters s
JOIN academic_years ay ON ay.id = s.academic_year_id
WHERE s.is_active = TRUE
ORDER BY s.start_date DESC
LIMIT 1;

-- name: CreateSemester :one
INSERT INTO semesters (id, academic_year_id, name, label, start_date, end_date, is_active)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: CountSemesterLabelConflicts :one
SELECT COUNT(*)::int
FROM semesters
WHERE LOWER(label) = LOWER(sqlc.arg(label));

-- name: ActivateSemester :one
UPDATE semesters
SET is_active = TRUE,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeactivateSemesters :exec
UPDATE semesters
SET is_active = FALSE,
    updated_at = NOW()
WHERE is_active = TRUE;

-- name: DeactivateSemesterAcademicYears :exec
UPDATE academic_years
SET is_active = FALSE,
    updated_at = NOW()
WHERE is_active = TRUE;

-- name: ActivateSemesterAcademicYear :one
UPDATE academic_years
SET is_active = TRUE,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSemester :exec
DELETE FROM semesters WHERE id = $1;
