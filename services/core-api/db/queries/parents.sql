-- name: ListParents :many
SELECT id, nama, phone, address, created_at, updated_at, photo_url, occupation, income_band, nik
FROM parents
ORDER BY nama ASC;

-- name: GetParent :one
SELECT id, nama, phone, address, created_at, updated_at, photo_url, occupation, income_band, nik
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
