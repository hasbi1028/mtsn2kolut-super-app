-- name: GetUserByUsername :one
SELECT id, username, password_hash, role, employee_id, created_at, updated_at
FROM users
WHERE username = $1;

-- name: ListUsers :many
SELECT u.id, u.username, u.role, u.employee_id, e.nama AS employee_nama, u.created_at
FROM users u
LEFT JOIN employees e ON e.id = u.employee_id
ORDER BY u.username ASC;

-- name: CreateUser :one
INSERT INTO users (username, password_hash, role, employee_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: CreateAuditLog :one
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListAuditLogs :many
SELECT a.id, a.user_id, u.username, a.action, a.entity_type, a.entity_id, a.metadata, a.created_at
FROM audit_logs a
LEFT JOIN users u ON u.id = a.user_id
ORDER BY a.created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteOldAuditLogs :execrows
DELETE FROM audit_logs
WHERE created_at < NOW() - INTERVAL '90 days';
