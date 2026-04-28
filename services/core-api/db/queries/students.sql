-- name: ListStudents :many
SELECT s.id, s.nis, s.nisn, s.nama, s.gender, s.parent_name, s.parent_phone,
       s.class_id, c.name AS class_name, c.code AS class_code,
       s.is_active, s.created_at, s.updated_at
FROM students s
LEFT JOIN school_classes c ON c.id = s.class_id
ORDER BY s.nama ASC;

-- name: CreateStudent :one
INSERT INTO students (id, nis, nisn, nama, gender, parent_name, parent_phone, class_id, is_active)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateStudent :one
UPDATE students
SET nis = $2, nisn = $3, nama = $4, gender = $5, 
    parent_name = $6, parent_phone = $7, class_id = $8, 
    is_active = $9, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteStudent :exec
DELETE FROM students WHERE id = $1;

-- name: ListStudentsByTeacher :many
SELECT DISTINCT
  s.id, s.nis, s.nisn, s.nama, s.gender, s.parent_name, s.parent_phone,
  s.class_id, c.name AS class_name, c.code AS class_code,
  s.is_active, s.created_at, s.updated_at
FROM students s
JOIN school_classes c ON c.id = s.class_id
JOIN class_subject_assignments csa ON csa.class_id = c.id
WHERE csa.teacher_employee_id = $1 AND s.is_active = TRUE
ORDER BY s.nama ASC;
