-- name: ListRombels :many
SELECT
    c.id,
    c.code,
    c.name,
    c.level,
    c.is_active,
    c.academic_year_id,
    ay.name AS academic_year_name,
    cha.id AS homeroom_assignment_id,
    cha.employee_id AS homeroom_employee_id,
    COALESCE(e.nama, '') AS homeroom_teacher_name,
    cha.start_date AS homeroom_start_date,
    cha.end_date AS homeroom_end_date,
    COUNT(DISTINCT s.id)::int AS total_students,
    COUNT(DISTINCT csa.teacher_employee_id)::int AS total_subject_teachers,
    COUNT(DISTINCT csa.id)::int AS total_subject_assignments,
    COUNT(DISTINCT ts.id)::int AS total_timetable_slots
FROM school_classes c
JOIN academic_years ay ON ay.id = c.academic_year_id
LEFT JOIN class_homeroom_assignments cha ON cha.class_id = c.id AND cha.is_active = TRUE
LEFT JOIN employees e ON e.id = cha.employee_id
LEFT JOIN students s ON s.class_id = c.id AND s.is_active = TRUE
LEFT JOIN class_subject_assignments csa ON csa.class_id = c.id
LEFT JOIN timetable_slots ts ON ts.assignment_id = csa.id
GROUP BY c.id, ay.name, ay.start_date, cha.id, cha.employee_id, e.nama, cha.start_date, cha.end_date
ORDER BY ay.start_date DESC, c.level ASC, c.name ASC;

-- name: GetRombelDetail :one
SELECT
    c.id,
    c.code,
    c.name,
    c.level,
    c.is_active,
    c.created_at,
    c.updated_at,
    c.academic_year_id,
    ay.name AS academic_year_name,
    ay.start_date AS academic_year_start_date,
    ay.end_date AS academic_year_end_date,
    cha.id AS homeroom_assignment_id,
    cha.employee_id AS homeroom_employee_id,
    COALESCE(e.nama, '') AS homeroom_teacher_name,
    cha.start_date AS homeroom_start_date,
    cha.end_date AS homeroom_end_date,
    COALESCE(cha.notes, '') AS homeroom_notes,
    COUNT(DISTINCT s.id)::int AS total_students,
    COUNT(DISTINCT ps.parent_id)::int AS total_linked_parents,
    COUNT(DISTINCT csa.teacher_employee_id)::int AS total_subject_teachers,
    COUNT(DISTINCT csa.id)::int AS total_subject_assignments,
    COUNT(DISTINCT ts.id)::int AS total_timetable_slots
FROM school_classes c
JOIN academic_years ay ON ay.id = c.academic_year_id
LEFT JOIN class_homeroom_assignments cha ON cha.class_id = c.id AND cha.is_active = TRUE
LEFT JOIN employees e ON e.id = cha.employee_id
LEFT JOIN students s ON s.class_id = c.id AND s.is_active = TRUE
LEFT JOIN parent_students ps ON ps.student_id = s.id
LEFT JOIN class_subject_assignments csa ON csa.class_id = c.id
LEFT JOIN timetable_slots ts ON ts.assignment_id = csa.id
WHERE c.id = $1
GROUP BY c.id, ay.name, ay.start_date, ay.end_date, cha.id, cha.employee_id, e.nama, cha.start_date, cha.end_date, cha.notes;

-- name: CountRombelCodeConflicts :one
SELECT COUNT(*)::int
FROM school_classes candidate
JOIN school_classes target ON target.id = sqlc.arg(id)
WHERE candidate.academic_year_id = target.academic_year_id
  AND candidate.id <> target.id
  AND LOWER(candidate.code) = LOWER(sqlc.arg(code));

-- name: UpdateRombelIdentity :one
WITH updated AS (
    UPDATE school_classes
    SET code = sqlc.arg(code),
        name = sqlc.arg(name),
        level = sqlc.arg(level),
        is_active = sqlc.arg(is_active),
        updated_at = NOW()
    WHERE school_classes.id = sqlc.arg(id)
    RETURNING *
)
SELECT
    c.id,
    c.code,
    c.name,
    c.level,
    c.is_active,
    c.created_at,
    c.updated_at,
    c.academic_year_id,
    ay.name AS academic_year_name,
    ay.start_date AS academic_year_start_date,
    ay.end_date AS academic_year_end_date,
    cha.id AS homeroom_assignment_id,
    cha.employee_id AS homeroom_employee_id,
    COALESCE(e.nama, '') AS homeroom_teacher_name,
    cha.start_date AS homeroom_start_date,
    cha.end_date AS homeroom_end_date,
    COALESCE(cha.notes, '') AS homeroom_notes,
    COUNT(DISTINCT s.id)::int AS total_students,
    COUNT(DISTINCT ps.parent_id)::int AS total_linked_parents,
    COUNT(DISTINCT csa.teacher_employee_id)::int AS total_subject_teachers,
    COUNT(DISTINCT csa.id)::int AS total_subject_assignments,
    COUNT(DISTINCT ts.id)::int AS total_timetable_slots
FROM updated c
JOIN academic_years ay ON ay.id = c.academic_year_id
LEFT JOIN class_homeroom_assignments cha ON cha.class_id = c.id AND cha.is_active = TRUE
LEFT JOIN employees e ON e.id = cha.employee_id
LEFT JOIN students s ON s.class_id = c.id AND s.is_active = TRUE
LEFT JOIN parent_students ps ON ps.student_id = s.id
LEFT JOIN class_subject_assignments csa ON csa.class_id = c.id
LEFT JOIN timetable_slots ts ON ts.assignment_id = csa.id
GROUP BY c.id, c.code, c.name, c.level, c.is_active, c.created_at, c.updated_at, c.academic_year_id, ay.name, ay.start_date, ay.end_date, cha.id, cha.employee_id, e.nama, cha.start_date, cha.end_date, cha.notes;

-- name: ListStudentsByClassWithParents :many
SELECT
    s.id AS student_id,
    s.nis,
    s.nisn,
    s.nama AS student_name,
    s.gender,
    s.parent_name,
    s.parent_phone,
    s.phone AS student_phone,
    s.alamat AS student_address,
    s.is_active,
    s.status,
    p.id AS parent_id,
    COALESCE(p.nama, '') AS parent_nama,
    COALESCE(p.phone, '') AS parent_phone_linked,
    COALESCE(p.address, '') AS parent_address,
    COALESCE(p.occupation, '') AS parent_occupation,
    COALESCE(p.income_band, '') AS parent_income_band,
    COALESCE(p.nik, '') AS parent_nik,
    COALESCE(ps.relationship::text, '') AS relationship,
    COALESCE(ps.is_primary_contact, FALSE) AS is_primary_contact,
    COALESCE(ps.notes, '') AS relationship_notes
FROM students s
LEFT JOIN parent_students ps ON ps.student_id = s.id
LEFT JOIN parents p ON p.id = ps.parent_id
WHERE s.class_id = $1
ORDER BY s.nama ASC, ps.is_primary_contact DESC, ps.relationship ASC, p.nama ASC;

-- name: ListRombelSubjectAssignments :many
SELECT
    csa.id,
    csa.class_id,
    c.name AS class_name,
    c.code AS class_code,
    csa.subject_id,
    sub.name AS subject_name,
    sub.code AS subject_code,
    csa.teacher_employee_id,
    e.nama AS teacher_name,
    csa.created_at,
    csa.updated_at
FROM class_subject_assignments csa
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects sub ON sub.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
WHERE csa.class_id = $1
ORDER BY sub.name ASC, e.nama ASC;

-- name: GetRombelSubjectAssignment :one
SELECT
    csa.id,
    csa.class_id,
    c.name AS class_name,
    c.code AS class_code,
    csa.subject_id,
    sub.name AS subject_name,
    sub.code AS subject_code,
    csa.teacher_employee_id,
    e.nama AS teacher_name,
    csa.created_at,
    csa.updated_at
FROM class_subject_assignments csa
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects sub ON sub.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
WHERE csa.class_id = sqlc.arg(class_id)
  AND csa.id = sqlc.arg(id);

-- name: CreateRombelSubjectAssignment :one
WITH inserted AS (
    INSERT INTO class_subject_assignments (id, class_id, subject_id, teacher_employee_id)
    VALUES (
        gen_random_uuid(),
        sqlc.arg(class_id),
        sqlc.arg(subject_id),
        sqlc.arg(teacher_employee_id)
    )
    RETURNING *
)
SELECT
    inserted.id,
    inserted.class_id,
    c.name AS class_name,
    c.code AS class_code,
    inserted.subject_id,
    sub.name AS subject_name,
    sub.code AS subject_code,
    inserted.teacher_employee_id,
    e.nama AS teacher_name,
    inserted.created_at,
    inserted.updated_at
FROM inserted
JOIN school_classes c ON c.id = inserted.class_id
JOIN subjects sub ON sub.id = inserted.subject_id
JOIN employees e ON e.id = inserted.teacher_employee_id;

-- name: UpdateRombelSubjectAssignment :one
WITH updated AS (
    UPDATE class_subject_assignments
    SET subject_id = sqlc.arg(subject_id),
        teacher_employee_id = sqlc.arg(teacher_employee_id),
        updated_at = NOW()
    WHERE class_subject_assignments.class_id = sqlc.arg(class_id)
      AND class_subject_assignments.id = sqlc.arg(id)
    RETURNING *
)
SELECT
    updated.id,
    updated.class_id,
    c.name AS class_name,
    c.code AS class_code,
    updated.subject_id,
    sub.name AS subject_name,
    sub.code AS subject_code,
    updated.teacher_employee_id,
    e.nama AS teacher_name,
    updated.created_at,
    updated.updated_at
FROM updated
JOIN school_classes c ON c.id = updated.class_id
JOIN subjects sub ON sub.id = updated.subject_id
JOIN employees e ON e.id = updated.teacher_employee_id;

-- name: CountRombelSubjectAssignmentDependents :one
SELECT
    (SELECT COUNT(*) FROM timetable_slots WHERE assignment_id = csa.id)::int AS total_timetable_slots,
    (SELECT COUNT(*) FROM class_journal_sessions WHERE assignment_id = csa.id)::int AS total_journal_sessions,
    (SELECT COUNT(*) FROM grade_components WHERE assignment_id = csa.id)::int AS total_grade_components,
    (SELECT COUNT(*) FROM grade_assignment_finalizations WHERE assignment_id = csa.id)::int AS total_grade_finalizations
FROM class_subject_assignments csa
WHERE csa.class_id = sqlc.arg(class_id)
  AND csa.id = sqlc.arg(id);

-- name: DeleteRombelSubjectAssignment :execrows
DELETE FROM class_subject_assignments
WHERE class_id = sqlc.arg(class_id)
  AND id = sqlc.arg(id);

-- name: ListRombelTimetableSlots :many
SELECT
    ts.id,
    ts.assignment_id,
    ts.day_of_week,
    ts.lesson_period_id,
    COALESCE(lpt.period_number, 0)::int AS period_number,
    ts.slot_type,
    ts.lesson_hours::float8 AS lesson_hours,
    COALESCE(lpt.label, '')::text AS lesson_period_label,
    COALESCE(lpt.activity_type, ts.slot_type)::text AS lesson_period_activity_type,
    COALESCE(lpt.is_counted_as_lesson, ts.lesson_hours > 0)::boolean AS is_counted_as_lesson,
    ts.start_time,
    ts.end_time,
    ts.room_label,
    ts.notes,
    ts.created_at,
    ts.updated_at,
    csa.class_id,
    c.name AS class_name,
    c.code AS class_code,
    csa.subject_id,
    sub.name AS subject_name,
    sub.code AS subject_code,
    csa.teacher_employee_id,
    e.nama AS teacher_name
FROM timetable_slots ts
JOIN class_subject_assignments csa ON csa.id = ts.assignment_id
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects sub ON sub.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
LEFT JOIN lesson_period_templates lpt ON lpt.id = ts.lesson_period_id
WHERE csa.class_id = $1
ORDER BY ts.day_of_week ASC, ts.start_time ASC, sub.name ASC;

-- name: GetRombelTimetableSlot :one
SELECT
    ts.id,
    ts.assignment_id,
    ts.day_of_week,
    ts.lesson_period_id,
    COALESCE(lpt.period_number, 0)::int AS period_number,
    ts.slot_type,
    ts.lesson_hours::float8 AS lesson_hours,
    COALESCE(lpt.label, '')::text AS lesson_period_label,
    COALESCE(lpt.activity_type, ts.slot_type)::text AS lesson_period_activity_type,
    COALESCE(lpt.is_counted_as_lesson, ts.lesson_hours > 0)::boolean AS is_counted_as_lesson,
    ts.start_time,
    ts.end_time,
    ts.room_label,
    ts.notes,
    ts.created_at,
    ts.updated_at,
    csa.class_id,
    c.name AS class_name,
    c.code AS class_code,
    csa.subject_id,
    sub.name AS subject_name,
    sub.code AS subject_code,
    csa.teacher_employee_id,
    e.nama AS teacher_name
FROM timetable_slots ts
JOIN class_subject_assignments csa ON csa.id = ts.assignment_id
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects sub ON sub.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
LEFT JOIN lesson_period_templates lpt ON lpt.id = ts.lesson_period_id
WHERE csa.class_id = sqlc.arg(class_id)
  AND ts.id = sqlc.arg(id);

-- name: CreateRombelTimetableSlot :one
WITH inserted AS (
    INSERT INTO timetable_slots (
        assignment_id, day_of_week, start_time, end_time, room_label, notes,
        lesson_period_id, slot_type, lesson_hours
    )
    SELECT
        sqlc.arg(assignment_id),
        sqlc.arg(day_of_week),
        sqlc.arg(start_time),
        sqlc.arg(end_time),
        sqlc.arg(room_label),
        sqlc.arg(notes),
        sqlc.narg(lesson_period_id),
        sqlc.arg(slot_type),
        sqlc.arg(lesson_hours)
    WHERE EXISTS (
        SELECT 1
        FROM class_subject_assignments csa
        WHERE csa.id = sqlc.arg(assignment_id)
          AND csa.class_id = sqlc.arg(class_id)
    )
      AND (
        sqlc.narg(lesson_period_id)::uuid IS NULL
        OR EXISTS (
          SELECT 1
          FROM lesson_period_templates lpt
          JOIN school_classes c ON c.academic_year_id = lpt.academic_year_id
          WHERE lpt.id = sqlc.narg(lesson_period_id)::uuid
            AND c.id = sqlc.arg(class_id)
            AND lpt.day_of_week = sqlc.arg(day_of_week)
        )
      )
    RETURNING *
)
SELECT
    inserted.id,
    inserted.assignment_id,
    inserted.day_of_week,
    inserted.lesson_period_id,
    COALESCE(lpt.period_number, 0)::int AS period_number,
    inserted.slot_type,
    inserted.lesson_hours::float8 AS lesson_hours,
    COALESCE(lpt.label, '')::text AS lesson_period_label,
    COALESCE(lpt.activity_type, inserted.slot_type)::text AS lesson_period_activity_type,
    COALESCE(lpt.is_counted_as_lesson, inserted.lesson_hours > 0)::boolean AS is_counted_as_lesson,
    inserted.start_time,
    inserted.end_time,
    inserted.room_label,
    inserted.notes,
    inserted.created_at,
    inserted.updated_at,
    csa.class_id,
    c.name AS class_name,
    c.code AS class_code,
    csa.subject_id,
    sub.name AS subject_name,
    sub.code AS subject_code,
    csa.teacher_employee_id,
    e.nama AS teacher_name
FROM inserted
JOIN class_subject_assignments csa ON csa.id = inserted.assignment_id
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects sub ON sub.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
LEFT JOIN lesson_period_templates lpt ON lpt.id = inserted.lesson_period_id;

-- name: UpdateRombelTimetableSlot :one
WITH updated AS (
    UPDATE timetable_slots
    SET assignment_id = sqlc.arg(assignment_id),
        day_of_week = sqlc.arg(day_of_week),
        start_time = sqlc.arg(start_time),
        end_time = sqlc.arg(end_time),
        room_label = sqlc.arg(room_label),
        notes = sqlc.arg(notes),
        lesson_period_id = sqlc.narg(lesson_period_id),
        slot_type = sqlc.arg(slot_type),
        lesson_hours = sqlc.arg(lesson_hours),
        updated_at = NOW()
    FROM class_subject_assignments current_assignment
    WHERE timetable_slots.id = sqlc.arg(id)
      AND current_assignment.id = timetable_slots.assignment_id
      AND current_assignment.class_id = sqlc.arg(class_id)
      AND EXISTS (
          SELECT 1
          FROM class_subject_assignments next_assignment
          WHERE next_assignment.id = sqlc.arg(assignment_id)
            AND next_assignment.class_id = sqlc.arg(class_id)
      )
      AND (
        sqlc.narg(lesson_period_id)::uuid IS NULL
        OR EXISTS (
          SELECT 1
          FROM lesson_period_templates lpt
          JOIN school_classes c ON c.academic_year_id = lpt.academic_year_id
          WHERE lpt.id = sqlc.narg(lesson_period_id)::uuid
            AND c.id = sqlc.arg(class_id)
            AND lpt.day_of_week = sqlc.arg(day_of_week)
        )
      )
    RETURNING timetable_slots.*
)
SELECT
    updated.id,
    updated.assignment_id,
    updated.day_of_week,
    updated.lesson_period_id,
    COALESCE(lpt.period_number, 0)::int AS period_number,
    updated.slot_type,
    updated.lesson_hours::float8 AS lesson_hours,
    COALESCE(lpt.label, '')::text AS lesson_period_label,
    COALESCE(lpt.activity_type, updated.slot_type)::text AS lesson_period_activity_type,
    COALESCE(lpt.is_counted_as_lesson, updated.lesson_hours > 0)::boolean AS is_counted_as_lesson,
    updated.start_time,
    updated.end_time,
    updated.room_label,
    updated.notes,
    updated.created_at,
    updated.updated_at,
    csa.class_id,
    c.name AS class_name,
    c.code AS class_code,
    csa.subject_id,
    sub.name AS subject_name,
    sub.code AS subject_code,
    csa.teacher_employee_id,
    e.nama AS teacher_name
FROM updated
JOIN class_subject_assignments csa ON csa.id = updated.assignment_id
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects sub ON sub.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
LEFT JOIN lesson_period_templates lpt ON lpt.id = updated.lesson_period_id;

-- name: DeleteRombelTimetableSlot :execrows
DELETE FROM timetable_slots ts
USING class_subject_assignments csa
WHERE csa.id = ts.assignment_id
  AND csa.class_id = sqlc.arg(class_id)
  AND ts.id = sqlc.arg(id);

-- name: ListHomeroomAssignmentsByClass :many
SELECT
    cha.id,
    cha.class_id,
    c.code AS class_code,
    c.name AS class_name,
    cha.employee_id,
    e.nama AS employee_name,
    cha.academic_year_id,
    COALESCE(ay.name, '') AS academic_year_name,
    cha.start_date,
    cha.end_date,
    cha.is_active,
    cha.notes,
    cha.created_at,
    cha.updated_at
FROM class_homeroom_assignments cha
JOIN school_classes c ON c.id = cha.class_id
JOIN employees e ON e.id = cha.employee_id
LEFT JOIN academic_years ay ON ay.id = cha.academic_year_id
WHERE cha.class_id = $1
ORDER BY cha.is_active DESC, cha.start_date DESC, cha.created_at DESC;

-- name: CreateHomeroomAssignment :one
WITH deactivate_existing AS (
    UPDATE class_homeroom_assignments
    SET is_active = FALSE,
        updated_at = NOW()
    WHERE class_homeroom_assignments.class_id = sqlc.arg(class_id)
      AND class_homeroom_assignments.is_active = TRUE
      AND sqlc.arg(homeroom_is_active)::boolean = TRUE
    RETURNING class_homeroom_assignments.id
), inserted AS (
    INSERT INTO class_homeroom_assignments (
        class_id, employee_id, academic_year_id, start_date, end_date, is_active, notes
    )
    VALUES (
        sqlc.arg(class_id),
        sqlc.arg(employee_id),
        COALESCE(sqlc.narg(homeroom_academic_year_id), (SELECT academic_year_id FROM school_classes WHERE id = sqlc.arg(class_id))),
        COALESCE(sqlc.narg(homeroom_start_date), CURRENT_DATE),
        sqlc.narg(homeroom_end_date),
        sqlc.arg(homeroom_is_active),
        sqlc.arg(notes)
    )
    RETURNING *
)
SELECT
    inserted.id,
    inserted.class_id,
    c.code AS class_code,
    c.name AS class_name,
    inserted.employee_id,
    e.nama AS employee_name,
    inserted.academic_year_id,
    COALESCE(ay.name, '') AS academic_year_name,
    inserted.start_date,
    inserted.end_date,
    inserted.is_active,
    inserted.notes,
    inserted.created_at,
    inserted.updated_at
FROM inserted
JOIN school_classes c ON c.id = inserted.class_id
JOIN employees e ON e.id = inserted.employee_id
LEFT JOIN academic_years ay ON ay.id = inserted.academic_year_id;

-- name: UpdateHomeroomAssignment :one
WITH target AS (
    SELECT cha.id, cha.class_id
    FROM class_homeroom_assignments cha
    WHERE cha.id = sqlc.arg(id)
), deactivate_existing AS (
    UPDATE class_homeroom_assignments
    SET is_active = FALSE,
        updated_at = NOW()
    WHERE class_homeroom_assignments.class_id = (SELECT class_id FROM target)
      AND class_homeroom_assignments.id <> sqlc.arg(id)
      AND class_homeroom_assignments.is_active = TRUE
      AND sqlc.arg(homeroom_is_active)::boolean = TRUE
    RETURNING class_homeroom_assignments.id
), updated AS (
    UPDATE class_homeroom_assignments
    SET employee_id = sqlc.arg(employee_id),
        is_active = sqlc.arg(homeroom_is_active),
        notes = sqlc.arg(notes),
        updated_at = NOW()
    WHERE class_homeroom_assignments.id = sqlc.arg(id)
    RETURNING *
)
SELECT
    updated.id,
    updated.class_id,
    c.code AS class_code,
    c.name AS class_name,
    updated.employee_id,
    e.nama AS employee_name,
    updated.academic_year_id,
    COALESCE(ay.name, '') AS academic_year_name,
    updated.start_date,
    updated.end_date,
    updated.is_active,
    updated.notes,
    updated.created_at,
    updated.updated_at
FROM updated
JOIN school_classes c ON c.id = updated.class_id
JOIN employees e ON e.id = updated.employee_id
LEFT JOIN academic_years ay ON ay.id = updated.academic_year_id;

-- name: DeleteHomeroomAssignment :exec
DELETE FROM class_homeroom_assignments
WHERE id = $1;

-- name: AssignStudentToClass :exec
UPDATE students
SET class_id = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: BulkAssignStudentsToClass :exec
UPDATE students
SET class_id = $2,
    updated_at = NOW()
WHERE id = ANY($1::uuid[]);

-- name: RemoveStudentFromClass :exec
UPDATE students
SET class_id = NULL,
    updated_at = NOW()
WHERE id = $1;

-- name: ListUnassignedStudents :many
SELECT s.id, s.nis, s.nisn, s.nama, s.gender,
       s.is_active, s.status
FROM students s
WHERE s.class_id IS NULL
  AND s.is_active = TRUE
ORDER BY s.nama ASC;
