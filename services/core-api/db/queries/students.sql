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

-- name: DeleteStudent :exec
DELETE FROM students WHERE id = $1;
