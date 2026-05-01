-- name: GetKesiswaanStats :one
SELECT
    COUNT(*) FILTER (WHERE s.status = 'active')::BIGINT AS active_students,
    COUNT(*) FILTER (WHERE s.status = 'prospective')::BIGINT AS prospective_students,
    COUNT(*) FILTER (WHERE s.total_violation_points > 0)::BIGINT AS students_with_points,
    COALESCE(SUM(s.total_violation_points), 0)::BIGINT AS total_violation_points,
    (SELECT COUNT(*) FROM student_violations sv WHERE sv.status = 'open')::BIGINT AS open_violations,
    (SELECT COUNT(*) FROM student_achievements sa WHERE EXTRACT(YEAR FROM sa.achievement_date) = EXTRACT(YEAR FROM NOW()))::BIGINT AS achievements_this_year,
    (SELECT COUNT(*) FROM violation_categories vc WHERE vc.is_active)::BIGINT AS active_categories,
    (SELECT COUNT(*) FROM extracurriculars e WHERE e.is_active)::BIGINT AS active_extracurriculars,
    (SELECT COUNT(*) FROM counseling_sessions cs WHERE cs.status IN ('open', 'monitoring'))::BIGINT AS open_counseling_sessions,
    (SELECT COUNT(*) FROM student_transfers st WHERE EXTRACT(YEAR FROM st.transfer_date) = EXTRACT(YEAR FROM NOW()) AND st.status = 'completed')::BIGINT AS transfers_this_year
FROM students s;

-- name: GetKesiswaanStatsByTeacher :one
SELECT
    COUNT(*) FILTER (WHERE s.status = 'active')::BIGINT AS active_students,
    COUNT(*) FILTER (WHERE s.status = 'prospective')::BIGINT AS prospective_students,
    COUNT(*) FILTER (WHERE s.total_violation_points > 0)::BIGINT AS students_with_points,
    COALESCE(SUM(s.total_violation_points), 0)::BIGINT AS total_violation_points,
    (
        SELECT COUNT(*)
        FROM student_violations sv
        JOIN students scoped ON scoped.id = sv.student_id
        WHERE sv.status = 'open'
          AND EXISTS (
              SELECT 1
              FROM class_subject_assignments csa
              WHERE csa.class_id = scoped.class_id
                AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
          )
    )::BIGINT AS open_violations,
    (
        SELECT COUNT(*)
        FROM student_achievements sa
        JOIN students scoped ON scoped.id = sa.student_id
        WHERE EXTRACT(YEAR FROM sa.achievement_date) = EXTRACT(YEAR FROM NOW())
          AND EXISTS (
              SELECT 1
              FROM class_subject_assignments csa
              WHERE csa.class_id = scoped.class_id
                AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
          )
    )::BIGINT AS achievements_this_year,
    (SELECT COUNT(*) FROM violation_categories vc WHERE vc.is_active)::BIGINT AS active_categories,
    (SELECT COUNT(*) FROM extracurriculars e WHERE e.is_active)::BIGINT AS active_extracurriculars,
    (
        SELECT COUNT(*)
        FROM counseling_sessions cs
        JOIN students scoped ON scoped.id = cs.student_id
        WHERE cs.status IN ('open', 'monitoring')
          AND cs.is_confidential = FALSE
          AND EXISTS (
              SELECT 1
              FROM class_subject_assignments csa
              WHERE csa.class_id = scoped.class_id
                AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
          )
    )::BIGINT AS open_counseling_sessions,
    (
        SELECT COUNT(*)
        FROM student_transfers st
        JOIN students scoped ON scoped.id = st.student_id
        WHERE EXTRACT(YEAR FROM st.transfer_date) = EXTRACT(YEAR FROM NOW())
          AND st.status = 'completed'
          AND EXISTS (
              SELECT 1
              FROM class_subject_assignments csa
              WHERE csa.class_id = scoped.class_id
                AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
          )
    )::BIGINT AS transfers_this_year
FROM students s
WHERE EXISTS (
    SELECT 1
    FROM class_subject_assignments csa
    WHERE csa.class_id = s.class_id
      AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
);

-- name: ListKesiswaanClassOptions :many
SELECT id, code, name, level
FROM school_classes
WHERE is_active = TRUE
ORDER BY level, name;

-- name: ListKesiswaanStudents :many
SELECT
    s.id, s.nis, s.nisn, s.nama, s.gender, s.parent_name, s.parent_phone,
    s.class_id, c.name AS class_name, c.code AS class_code,
    s.is_active, s.status, s.nik, s.tempat_lahir, s.tanggal_lahir, s.alamat,
    s.agama, s.anak_ke, s.phone, s.photo_url, s.total_violation_points,
    COALESCE(v.violation_count, 0)::BIGINT AS violation_count,
    COALESCE(a.achievement_count, 0)::BIGINT AS achievement_count,
    s.created_at, s.updated_at
FROM students s
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS violation_count
    FROM student_violations sv
    WHERE sv.student_id = s.id AND sv.status <> 'canceled'
) v ON TRUE
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS achievement_count
    FROM student_achievements sa
    WHERE sa.student_id = s.id
) a ON TRUE
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    s.nis ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nisn ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nama ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nik ILIKE '%' || sqlc.arg(search) || '%' OR
    c.name ILIKE '%' || sqlc.arg(search) || '%' OR
    c.code ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(status)::TEXT = '' OR s.status::TEXT = sqlc.arg(status)
) AND (
    sqlc.arg(class_id)::UUID IS NULL OR s.class_id = sqlc.arg(class_id)::UUID
) AND (
    sqlc.arg(teacher_employee_id)::UUID IS NULL OR EXISTS (
        SELECT 1
        FROM class_subject_assignments csa
        WHERE csa.class_id = s.class_id
          AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
    )
)
ORDER BY c.level, c.name, s.nama;

-- name: UpdateKesiswaanStudentProfile :one
UPDATE students
SET nik = $2,
    tempat_lahir = $3,
    tanggal_lahir = $4,
    alamat = $5,
    agama = $6,
    anak_ke = $7,
    phone = $8,
    parent_name = $9,
    parent_phone = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateKesiswaanStudentPhoto :one
UPDATE students
SET photo_url = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListViolationCategories :many
SELECT *
FROM violation_categories
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    code ILIKE '%' || sqlc.arg(search) || '%' OR
    name ILIKE '%' || sqlc.arg(search) || '%' OR
    description ILIKE '%' || sqlc.arg(search) || '%'
)
ORDER BY is_active DESC, severity, code;

-- name: CreateViolationCategory :one
INSERT INTO violation_categories (code, name, point, severity, description, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateViolationCategory :one
UPDATE violation_categories
SET code = $2,
    name = $3,
    point = $4,
    severity = $5,
    description = $6,
    is_active = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteViolationCategory :exec
DELETE FROM violation_categories WHERE id = $1;

-- name: ListStudentViolations :many
SELECT
    sv.id, sv.student_id, sv.category_id, sv.incident_date, sv.points,
    sv.description, sv.action_taken, sv.status, sv.reported_by_employee_id,
    sv.recorded_by_user_id, sv.created_at, sv.updated_at,
    s.nama AS student_name,
    s.nis AS student_nis,
    c.name AS class_name,
    c.code AS class_code,
    vc.code AS category_code,
    vc.name AS category_name,
    e.nama AS reported_by_name
FROM student_violations sv
JOIN students s ON s.id = sv.student_id
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN violation_categories vc ON vc.id = sv.category_id
LEFT JOIN employees e ON e.id = sv.reported_by_employee_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    s.nama ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nis ILIKE '%' || sqlc.arg(search) || '%' OR
    sv.description ILIKE '%' || sqlc.arg(search) || '%' OR
    sv.action_taken ILIKE '%' || sqlc.arg(search) || '%' OR
    vc.name ILIKE '%' || sqlc.arg(search) || '%' OR
    vc.code ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(status)::TEXT = '' OR sv.status = sqlc.arg(status)
) AND (
    sqlc.arg(student_id)::UUID IS NULL OR sv.student_id = sqlc.arg(student_id)::UUID
) AND (
    sqlc.arg(teacher_employee_id)::UUID IS NULL OR EXISTS (
        SELECT 1
        FROM class_subject_assignments csa
        WHERE csa.class_id = s.class_id
          AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
    )
)
ORDER BY sv.incident_date DESC, sv.created_at DESC;

-- name: CreateStudentViolation :one
INSERT INTO student_violations (
    student_id, category_id, incident_date, points, description, action_taken,
    status, reported_by_employee_id, recorded_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateStudentViolation :one
UPDATE student_violations
SET student_id = $2,
    category_id = $3,
    incident_date = $4,
    points = $5,
    description = $6,
    action_taken = $7,
    status = $8,
    reported_by_employee_id = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteStudentViolation :exec
DELETE FROM student_violations WHERE id = $1;

-- name: ListStudentAchievements :many
SELECT
    sa.id, sa.student_id, sa.achievement_date, sa.title, sa.level, sa.category,
    sa.organizer, sa.description, sa.document_url, sa.recorded_by_user_id,
    sa.created_at, sa.updated_at,
    s.nama AS student_name,
    s.nis AS student_nis,
    c.name AS class_name,
    c.code AS class_code
FROM student_achievements sa
JOIN students s ON s.id = sa.student_id
LEFT JOIN school_classes c ON c.id = s.class_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    s.nama ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nis ILIKE '%' || sqlc.arg(search) || '%' OR
    sa.title ILIKE '%' || sqlc.arg(search) || '%' OR
    sa.category ILIKE '%' || sqlc.arg(search) || '%' OR
    sa.organizer ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(level)::TEXT = '' OR sa.level = sqlc.arg(level)
) AND (
    sqlc.arg(student_id)::UUID IS NULL OR sa.student_id = sqlc.arg(student_id)::UUID
) AND (
    sqlc.arg(teacher_employee_id)::UUID IS NULL OR EXISTS (
        SELECT 1
        FROM class_subject_assignments csa
        WHERE csa.class_id = s.class_id
          AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
    )
)
ORDER BY sa.achievement_date DESC, sa.created_at DESC;

-- name: CreateStudentAchievement :one
INSERT INTO student_achievements (
    student_id, achievement_date, title, level, category, organizer,
    description, document_url, recorded_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateStudentAchievement :one
UPDATE student_achievements
SET student_id = $2,
    achievement_date = $3,
    title = $4,
    level = $5,
    category = $6,
    organizer = $7,
    description = $8,
    document_url = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteStudentAchievement :exec
DELETE FROM student_achievements WHERE id = $1;

-- name: ListExtracurriculars :many
SELECT
    e.id, e.code, e.name, e.category, e.description, e.supervisor_employee_id,
    e.schedule_text, e.is_active, e.created_at, e.updated_at,
    COALESCE(emp.nama, '')::TEXT AS supervisor_name,
    COALESCE(m.active_member_count, 0)::BIGINT AS active_member_count
FROM extracurriculars e
LEFT JOIN employees emp ON emp.id = e.supervisor_employee_id
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS active_member_count
    FROM extracurricular_members em
    WHERE em.extracurricular_id = e.id AND em.status = 'active'
) m ON TRUE
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    e.code ILIKE '%' || sqlc.arg(search) || '%' OR
    e.name ILIKE '%' || sqlc.arg(search) || '%' OR
    e.category ILIKE '%' || sqlc.arg(search) || '%' OR
    e.description ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(active_only)::BOOLEAN = FALSE OR e.is_active = TRUE
)
ORDER BY e.is_active DESC, e.name;

-- name: CreateExtracurricular :one
INSERT INTO extracurriculars (
    code, name, category, description, supervisor_employee_id, schedule_text, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateExtracurricular :one
UPDATE extracurriculars
SET code = $2,
    name = $3,
    category = $4,
    description = $5,
    supervisor_employee_id = $6,
    schedule_text = $7,
    is_active = $8,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteExtracurricular :exec
DELETE FROM extracurriculars WHERE id = $1;

-- name: ListExtracurricularMembers :many
SELECT
    em.id, em.extracurricular_id, em.student_id, em.joined_at, em.role,
    em.status, em.notes, em.recorded_by_user_id, em.created_at, em.updated_at,
    e.code AS extracurricular_code,
    e.name AS extracurricular_name,
    s.nama AS student_name,
    s.nis AS student_nis,
    c.name AS class_name,
    c.code AS class_code
FROM extracurricular_members em
JOIN extracurriculars e ON e.id = em.extracurricular_id
JOIN students s ON s.id = em.student_id
LEFT JOIN school_classes c ON c.id = s.class_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    e.code ILIKE '%' || sqlc.arg(search) || '%' OR
    e.name ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nama ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nis ILIKE '%' || sqlc.arg(search) || '%' OR
    em.notes ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(status)::TEXT = '' OR em.status = sqlc.arg(status)
) AND (
    sqlc.arg(extracurricular_id)::UUID IS NULL OR em.extracurricular_id = sqlc.arg(extracurricular_id)::UUID
) AND (
    sqlc.arg(student_id)::UUID IS NULL OR em.student_id = sqlc.arg(student_id)::UUID
) AND (
    sqlc.arg(teacher_employee_id)::UUID IS NULL OR EXISTS (
        SELECT 1
        FROM class_subject_assignments csa
        WHERE csa.class_id = s.class_id
          AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
    )
)
ORDER BY e.name, c.level, c.name, s.nama;

-- name: CreateExtracurricularMember :one
INSERT INTO extracurricular_members (
    extracurricular_id, student_id, joined_at, role, status, notes, recorded_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateExtracurricularMember :one
UPDATE extracurricular_members
SET extracurricular_id = $2,
    student_id = $3,
    joined_at = $4,
    role = $5,
    status = $6,
    notes = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteExtracurricularMember :exec
DELETE FROM extracurricular_members WHERE id = $1;

-- name: ListCounselingSessions :many
SELECT
    cs.id, cs.student_id, cs.session_date, cs.topic, cs.summary, cs.follow_up,
    cs.status, cs.is_confidential, cs.counselor_employee_id, cs.recorded_by_user_id,
    cs.created_at, cs.updated_at,
    s.nama AS student_name,
    s.nis AS student_nis,
    c.name AS class_name,
    c.code AS class_code,
    COALESCE(emp.nama, '')::TEXT AS counselor_name
FROM counseling_sessions cs
JOIN students s ON s.id = cs.student_id
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN employees emp ON emp.id = cs.counselor_employee_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    s.nama ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nis ILIKE '%' || sqlc.arg(search) || '%' OR
    cs.topic ILIKE '%' || sqlc.arg(search) || '%' OR
    cs.summary ILIKE '%' || sqlc.arg(search) || '%' OR
    cs.follow_up ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(status)::TEXT = '' OR cs.status = sqlc.arg(status)
) AND (
    sqlc.arg(student_id)::UUID IS NULL OR cs.student_id = sqlc.arg(student_id)::UUID
) AND (
    sqlc.arg(can_read_confidential)::BOOLEAN = TRUE OR cs.is_confidential = FALSE
) AND (
    sqlc.arg(teacher_employee_id)::UUID IS NULL OR EXISTS (
        SELECT 1
        FROM class_subject_assignments csa
        WHERE csa.class_id = s.class_id
          AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
    )
)
ORDER BY cs.session_date DESC, cs.created_at DESC;

-- name: CreateCounselingSession :one
INSERT INTO counseling_sessions (
    student_id, session_date, topic, summary, follow_up, status,
    is_confidential, counselor_employee_id, recorded_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateCounselingSession :one
UPDATE counseling_sessions
SET student_id = $2,
    session_date = $3,
    topic = $4,
    summary = $5,
    follow_up = $6,
    status = $7,
    is_confidential = $8,
    counselor_employee_id = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCounselingSession :exec
DELETE FROM counseling_sessions WHERE id = $1;

-- name: ListStudentTransfers :many
SELECT
    st.id, st.student_id, st.transfer_date, st.transfer_type, st.previous_school,
    st.destination_school, st.reason, st.document_ref, st.notes, st.status,
    st.recorded_by_user_id, st.created_at, st.updated_at,
    s.nama AS student_name,
    s.nis AS student_nis,
    s.status AS student_status,
    c.name AS class_name,
    c.code AS class_code
FROM student_transfers st
JOIN students s ON s.id = st.student_id
LEFT JOIN school_classes c ON c.id = s.class_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    s.nama ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nis ILIKE '%' || sqlc.arg(search) || '%' OR
    st.previous_school ILIKE '%' || sqlc.arg(search) || '%' OR
    st.destination_school ILIKE '%' || sqlc.arg(search) || '%' OR
    st.reason ILIKE '%' || sqlc.arg(search) || '%' OR
    st.document_ref ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(transfer_type)::TEXT = '' OR st.transfer_type = sqlc.arg(transfer_type)
) AND (
    sqlc.arg(student_id)::UUID IS NULL OR st.student_id = sqlc.arg(student_id)::UUID
) AND (
    sqlc.arg(teacher_employee_id)::UUID IS NULL OR EXISTS (
        SELECT 1
        FROM class_subject_assignments csa
        WHERE csa.class_id = s.class_id
          AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
    )
)
ORDER BY st.transfer_date DESC, st.created_at DESC;

-- name: CreateStudentTransfer :one
INSERT INTO student_transfers (
    student_id, transfer_date, transfer_type, previous_school, destination_school,
    reason, document_ref, notes, status, recorded_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'completed', $9)
RETURNING *;

-- name: UpdateKesiswaanStudentLifecycle :one
UPDATE students
SET status = $2,
    is_active = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
