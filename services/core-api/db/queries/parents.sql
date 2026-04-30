-- name: ListParents :many
SELECT id, nama, phone, address, created_at, updated_at
FROM parents
ORDER BY nama ASC;

-- name: GetParent :one
SELECT id, nama, phone, address, created_at, updated_at
FROM parents
WHERE id = $1;

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

-- name: UnlinkParentStudent :exec
DELETE FROM parent_students
WHERE parent_id = $1 AND student_id = $2;

-- name: ListParentChildren :many
SELECT s.id, s.nis, s.nama, s.class_id, c.name as class_name
FROM students s
JOIN parent_students ps ON ps.student_id = s.id
LEFT JOIN school_classes c ON c.id = s.class_id
WHERE ps.parent_id = $1;

-- name: ListStudentParents :many
SELECT p.id, p.nama, p.phone
FROM parents p
JOIN parent_students ps ON ps.parent_id = p.id
WHERE ps.student_id = $1;
