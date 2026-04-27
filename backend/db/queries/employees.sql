-- name: ListEmployees :many
SELECT id, nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active, created_at, updated_at
FROM employees
ORDER BY nama ASC;

-- name: ListActiveEmployees :many
SELECT id, nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active, created_at, updated_at
FROM employees
WHERE is_active = TRUE
ORDER BY nama ASC;

-- name: GetEmployee :one
SELECT id, nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active, created_at, updated_at
FROM employees
WHERE id = $1;

-- name: CreateEmployee :one
INSERT INTO employees (id, nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateEmployee :one
UPDATE employees
SET nip             = $2,
    nama            = $3,
    unit_kerja      = $4,
    pusaka_username = $5,
    pusaka_password = $6,
    is_active       = $7,
    updated_at      = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEmployee :exec
DELETE FROM employees WHERE id = $1;

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees;
