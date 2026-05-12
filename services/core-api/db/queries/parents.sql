-- name: ListParents :many
SELECT id, nama, phone, address, created_at, updated_at, photo_url, occupation, income_band, nik
FROM parents
ORDER BY nama ASC;

-- name: GetParent :one
SELECT id, nama, phone, address, created_at, updated_at, photo_url, occupation, income_band, nik
FROM parents
WHERE id = $1;

-- name: GetPortalParentIDByUserID :one
SELECT u.parent_id
FROM users u
WHERE u.id = $1
  AND u.deleted_at IS NULL
  AND u.is_active = TRUE
  AND u.parent_id IS NOT NULL;

-- name: ListParentPortalPreviewParents :many
SELECT p.id, p.nama, p.phone, COUNT(ps.student_id)::bigint AS linked_student_count
FROM parents p
LEFT JOIN parent_students ps ON ps.parent_id = p.id
GROUP BY p.id, p.nama, p.phone
ORDER BY p.nama ASC;

-- name: CreateParent :one
INSERT INTO parents (nama, phone, address)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateParent :one
UPDATE parents
SET nama = $2, phone = $3, address = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteParent :exec
DELETE FROM parents WHERE id = $1;

-- name: LinkParentStudent :exec
INSERT INTO parent_students (parent_id, student_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UpsertParentStudentRelationship :one
INSERT INTO parent_students (parent_id, student_id, relationship, is_primary_contact, notes)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (parent_id, student_id) DO UPDATE
SET relationship = EXCLUDED.relationship,
    is_primary_contact = EXCLUDED.is_primary_contact,
    notes = EXCLUDED.notes,
    updated_at = NOW()
RETURNING parent_id, student_id, relationship, is_primary_contact, notes, created_at, updated_at;

-- name: UnlinkParentStudent :exec
DELETE FROM parent_students
WHERE parent_id = $1 AND student_id = $2;

-- name: ListParentChildren :many
SELECT s.id, s.nis, s.nama, s.class_id, c.name as class_name,
       ps.relationship, ps.is_primary_contact, ps.notes
FROM students s
JOIN parent_students ps ON ps.student_id = s.id
LEFT JOIN school_classes c ON c.id = s.class_id
WHERE ps.parent_id = $1;

-- name: GetParentPortalChildAccess :one
SELECT ps.student_id
FROM parent_students ps
WHERE ps.parent_id = sqlc.arg(parent_id)
  AND ps.student_id = sqlc.arg(student_id);

-- name: GetParentPortalChildProfile :one
SELECT s.id, s.nis, s.nisn, s.nama, s.gender, s.parent_name, s.parent_phone,
       s.class_id, c.name AS class_name, c.code AS class_code,
       COALESCE(lp.linked_parent_names, '') AS linked_parent_names,
       COALESCE(lp.linked_parent_count, 0) AS linked_parent_count,
       s.is_active, s.status, s.created_at, s.updated_at
FROM students s
JOIN parent_students parent_scope
  ON parent_scope.student_id = s.id
 AND parent_scope.parent_id = sqlc.arg(parent_id)
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN LATERAL (
  SELECT STRING_AGG(p.nama, ', ' ORDER BY p.nama) AS linked_parent_names,
         COUNT(*)::bigint AS linked_parent_count
  FROM parent_students ps
  JOIN parents p ON p.id = ps.parent_id
  WHERE ps.student_id = s.id
) lp ON TRUE
WHERE s.id = sqlc.arg(student_id);

-- name: ListParentPortalChildTimetable :many
SELECT ts.id, ts.assignment_id, ts.day_of_week, ts.start_time, ts.end_time, ts.room_label, ts.notes,
       c.id AS class_id, c.name AS class_name, c.code AS class_code,
       s.id AS subject_id, s.name AS subject_name, s.code AS subject_code,
       e.id AS teacher_employee_id, e.nama AS teacher_name
FROM parent_students parent_scope
JOIN students st ON st.id = parent_scope.student_id
JOIN school_classes c ON c.id = st.class_id
JOIN class_subject_assignments a ON a.class_id = c.id
JOIN timetable_slots ts ON ts.assignment_id = a.id
JOIN subjects s ON s.id = a.subject_id
JOIN employees e ON e.id = a.teacher_employee_id
WHERE parent_scope.parent_id = sqlc.arg(parent_id)
  AND parent_scope.student_id = sqlc.arg(student_id)
ORDER BY ts.day_of_week ASC, ts.start_time ASC, s.name ASC;

-- name: ListParentPortalChildExamSessions :many
SELECT
  ep.id AS participant_id,
  ep.session_id,
  ep.room_id,
  ep.seat_no,
  ep.joined_at,
  ep.submitted_at,
  CASE WHEN s.status = 'finished' THEN ep.score ELSE NULL::numeric END AS score,
  s.title AS session_title,
  s.status AS session_status,
  s.scheduled_start,
  s.scheduled_end,
  p.title AS package_title,
  p.duration_minutes,
  COALESCE(r.room_name, '') AS room_name
FROM parent_students parent_scope
JOIN cbt_exam_participants ep ON ep.student_id = parent_scope.student_id
JOIN cbt_exam_sessions s ON s.id = ep.session_id
JOIN cbt_packages p ON p.id = s.package_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE parent_scope.parent_id = sqlc.arg(parent_id)
  AND parent_scope.student_id = sqlc.arg(student_id)
ORDER BY s.scheduled_start DESC;

-- name: ListStudentParents :many
SELECT p.id, p.nama, p.phone, p.address, p.occupation, p.income_band, p.nik,
       ps.relationship, ps.is_primary_contact, ps.notes
FROM parents p
JOIN parent_students ps ON ps.parent_id = p.id
WHERE ps.student_id = $1;

-- name: FindParentForNormalization :one
SELECT id, nama, phone, address, created_at, updated_at, photo_url, occupation, income_band, nik
FROM parents
WHERE LOWER(REGEXP_REPLACE(BTRIM(nama), '\s+', ' ', 'g')) = LOWER(REGEXP_REPLACE(BTRIM(sqlc.arg(nama)), '\s+', ' ', 'g'))
  AND (
    BTRIM(COALESCE(sqlc.arg(phone), '')) = ''
    OR phone = ''
    OR REGEXP_REPLACE(phone, '\D', '', 'g') = REGEXP_REPLACE(sqlc.arg(phone), '\D', '', 'g')
  )
  AND (
    BTRIM(COALESCE(sqlc.arg(address), '')) = ''
    OR address = ''
    OR LOWER(REGEXP_REPLACE(BTRIM(address), '\s+', ' ', 'g')) = LOWER(REGEXP_REPLACE(BTRIM(sqlc.arg(address)), '\s+', ' ', 'g'))
  )
ORDER BY
  CASE WHEN REGEXP_REPLACE(phone, '\D', '', 'g') <> '' AND REGEXP_REPLACE(phone, '\D', '', 'g') = REGEXP_REPLACE(sqlc.arg(phone), '\D', '', 'g') THEN 0 ELSE 1 END,
  CASE WHEN address <> '' AND LOWER(REGEXP_REPLACE(BTRIM(address), '\s+', ' ', 'g')) = LOWER(REGEXP_REPLACE(BTRIM(sqlc.arg(address)), '\s+', ' ', 'g')) THEN 0 ELSE 1 END,
  created_at ASC
LIMIT 1;

-- name: CreateParentFromNormalization :one
INSERT INTO parents (nama, phone, address)
VALUES ($1, $2, $3)
RETURNING *;
