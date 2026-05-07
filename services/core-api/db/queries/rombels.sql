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

-- name: ListRombelTimetableSlots :many
SELECT
    ts.id,
    ts.assignment_id,
    ts.day_of_week,
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
WHERE csa.class_id = $1
ORDER BY ts.day_of_week ASC, ts.start_time ASC, sub.name ASC;

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
