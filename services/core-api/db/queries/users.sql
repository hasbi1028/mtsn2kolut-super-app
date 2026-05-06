-- name: GetUserByUsername :one
SELECT 
    u.id, u.username, u.password_hash, 
    u.employee_id, u.student_id, u.parent_id,
    u.is_active, u.auth_version, u.created_at, u.updated_at,
    (SELECT json_agg(role) FROM user_account_roles WHERE user_id = u.id) as roles
FROM users u
WHERE u.username = $1;

-- name: GetUserByID :one
SELECT 
    u.id, u.username, u.password_hash, 
    u.employee_id, u.student_id, u.parent_id,
    u.is_active, u.auth_version, u.created_at, u.updated_at,
    (SELECT json_agg(role) FROM user_account_roles WHERE user_id = u.id) as roles
FROM users u
WHERE u.id = $1;

-- name: ListUsers :many
SELECT 
    u.id, u.username, u.employee_id, u.student_id, u.parent_id,
    COALESCE(e.nama, s.nama, p.nama, '') AS profile_nama,
    u.is_active, u.created_at,
    (SELECT json_agg(role) FROM user_account_roles WHERE user_id = u.id) as roles
FROM users u
LEFT JOIN employees e ON e.id = u.employee_id
LEFT JOIN students s ON s.id = u.student_id
LEFT JOIN parents p ON p.id = u.parent_id
ORDER BY u.username ASC;

-- name: CreateUser :one
INSERT INTO users (username, password_hash, employee_id, student_id, parent_id, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, username, employee_id, student_id, parent_id, is_active, auth_version, created_at, updated_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: IncrementUserAuthVersion :one
UPDATE users
SET auth_version = auth_version + 1,
    updated_at = NOW()
WHERE id = $1
RETURNING auth_version;

-- name: UpdateUserStatus :exec
UPDATE users
SET is_active = $2,
    auth_version = CASE WHEN is_active = TRUE AND $2 = FALSE THEN auth_version + 1 ELSE auth_version END,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserProfileLink :exec
UPDATE users
SET employee_id = $2,
    student_id = $3,
    parent_id = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: GetUserRoles :many
SELECT role FROM user_account_roles WHERE user_id = $1;

-- name: AddUserRole :exec
INSERT INTO user_account_roles (user_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: ListUsersByStudentID :many
SELECT id, username, password_hash, employee_id, created_at, updated_at, student_id, parent_id, is_active
FROM users
WHERE student_id = $1
ORDER BY created_at ASC;

-- name: ListUsersByEmployeeID :many
SELECT id, username, password_hash, employee_id, created_at, updated_at, student_id, parent_id, is_active
FROM users
WHERE employee_id = $1
ORDER BY created_at ASC;

-- name: RemoveUserRole :exec
DELETE FROM user_account_roles WHERE user_id = $1 AND role = $2;

-- name: RemoveAllUserRoles :exec
DELETE FROM user_account_roles WHERE user_id = $1;

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

-- name: ListEntityAuditLogs :many
SELECT a.id, a.user_id, u.username, a.action, a.entity_type, a.entity_id, a.metadata, a.created_at
FROM audit_logs a
LEFT JOIN users u ON u.id = a.user_id
WHERE (
    a.entity_type = $1
    AND a.entity_id = $2
  )
  OR (
    $1 = 'cbt_session'
    AND a.entity_type = 'cbt_session_room'
    AND a.metadata->>'session_id' = $2
  )
ORDER BY a.created_at DESC
LIMIT $3 OFFSET $4;

-- name: DeleteOldAuditLogs :execrows
DELETE FROM audit_logs
WHERE created_at < NOW() - INTERVAL '90 days';
-- name: ListEmployeeAccountGenerationCandidates :many
WITH active_employees AS (
    SELECT
        e.id,
        e.nip,
        e.nama,
        e.tanggal_lahir,
        row_number() OVER (ORDER BY e.nama ASC, e.id ASC)::int AS nomor_urut
    FROM employees e
    WHERE e.is_active = TRUE
), candidates AS (
    SELECT
        ae.id AS employee_id,
        ae.nip,
        ae.nama,
        ae.tanggal_lahir,
        ae.nomor_urut,
        CASE
            WHEN ae.tanggal_lahir IS NULL THEN ''::text
            ELSE (sqlc.arg(npsn)::text || to_char(ae.tanggal_lahir, 'DDMMYY') || lpad(ae.nomor_urut::text, 3, '0'))::text
        END AS generated_username,
        linked.id AS existing_user_id
    FROM active_employees ae
    LEFT JOIN users linked ON linked.employee_id = ae.id
)
SELECT
    c.employee_id,
    c.nip,
    c.nama,
    c.tanggal_lahir,
    c.nomor_urut,
    c.generated_username,
    c.existing_user_id,
    username_user.id AS username_user_id
FROM candidates c
LEFT JOIN users username_user ON username_user.username = c.generated_username AND c.generated_username <> ''
ORDER BY c.nomor_urut ASC;
