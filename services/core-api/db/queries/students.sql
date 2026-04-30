-- name: ListStudents :many
SELECT s.id, s.nis, s.nisn, s.nama, s.gender, s.parent_name, s.parent_phone,
       s.class_id, c.name AS class_name, c.code AS class_code,
       COALESCE(lp.linked_parent_names, '') AS linked_parent_names,
       COALESCE(lp.linked_parent_count, 0) AS linked_parent_count,
       s.is_active, s.status, s.created_at, s.updated_at
FROM students s
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN LATERAL (
  SELECT STRING_AGG(p.nama, ', ' ORDER BY p.nama) AS linked_parent_names,
         COUNT(*)::bigint AS linked_parent_count
  FROM parent_students ps
  JOIN parents p ON p.id = ps.parent_id
  WHERE ps.student_id = s.id
) lp ON TRUE
ORDER BY s.nama ASC;

-- name: GetStudentByID :one
SELECT s.id, s.nis, s.nisn, s.nama, s.gender, s.parent_name, s.parent_phone,
       s.class_id, c.name AS class_name, c.code AS class_code,
       COALESCE(lp.linked_parent_names, '') AS linked_parent_names,
       COALESCE(lp.linked_parent_count, 0) AS linked_parent_count,
       s.is_active, s.status, s.created_at, s.updated_at
FROM students s
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN LATERAL (
  SELECT STRING_AGG(p.nama, ', ' ORDER BY p.nama) AS linked_parent_names,
         COUNT(*)::bigint AS linked_parent_count
  FROM parent_students ps
  JOIN parents p ON p.id = ps.parent_id
  WHERE ps.student_id = s.id
) lp ON TRUE
WHERE s.id = $1;

-- name: CreateStudent :one
INSERT INTO students (nis, nisn, nama, gender, parent_name, parent_phone, class_id, is_active, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateStudent :one
UPDATE students
SET nis = $2, nisn = $3, nama = $4, gender = $5, 
    parent_name = $6, parent_phone = $7, class_id = $8, 
    is_active = $9, status = $10, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteStudent :exec
DELETE FROM students WHERE id = $1;

-- name: ListStudentsByTeacher :many
SELECT DISTINCT
  s.id, s.nis, s.nisn, s.nama, s.gender, s.parent_name, s.parent_phone,
  s.class_id, c.name AS class_name, c.code AS class_code,
  COALESCE(lp.linked_parent_names, '') AS linked_parent_names,
  COALESCE(lp.linked_parent_count, 0) AS linked_parent_count,
  s.is_active, s.status, s.created_at, s.updated_at
FROM students s
JOIN school_classes c ON c.id = s.class_id
JOIN class_subject_assignments csa ON csa.class_id = c.id
LEFT JOIN LATERAL (
  SELECT STRING_AGG(p.nama, ', ' ORDER BY p.nama) AS linked_parent_names,
         COUNT(*)::bigint AS linked_parent_count
  FROM parent_students ps
  JOIN parents p ON p.id = ps.parent_id
  WHERE ps.student_id = s.id
) lp ON TRUE
WHERE csa.teacher_employee_id = $1 AND s.is_active = TRUE
ORDER BY s.nama ASC;

-- name: UpdateStudentStatus :exec
UPDATE students SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateStudentLifecycle :exec
UPDATE students
SET status = $2,
    is_active = $3,
    updated_at = NOW()
WHERE id = $1;
