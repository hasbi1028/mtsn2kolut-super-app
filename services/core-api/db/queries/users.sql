-- name: GetUserByUsername :one
SELECT 
    u.id, u.username, u.password_hash,
    u.display_name,
    u.employee_id, u.student_id, u.parent_id,
    u.is_active, u.auth_version, u.last_login_at, u.deleted_at, u.created_at, u.updated_at,
    (SELECT json_agg(role) FROM user_account_roles WHERE user_id = u.id) as roles
FROM users u
WHERE u.username = $1
  AND u.deleted_at IS NULL;

-- name: GetUserByID :one
SELECT 
    u.id, u.username, u.password_hash,
    u.display_name,
    u.employee_id, u.student_id, u.parent_id,
    u.is_active, u.auth_version, u.last_login_at, u.deleted_at, u.created_at, u.updated_at,
    (SELECT json_agg(role) FROM user_account_roles WHERE user_id = u.id) as roles
FROM users u
WHERE u.id = $1
  AND u.deleted_at IS NULL;

-- name: ListUsers :many
SELECT 
    u.id, u.username,
    COALESCE(NULLIF(u.display_name, ''), e.nama, s.nama, p.nama, u.username) AS display_name,
    u.employee_id, u.student_id, u.parent_id,
    COALESCE(e.nama, s.nama, p.nama, '') AS profile_nama,
    u.is_active, u.last_login_at, u.deleted_at, u.created_at,
    (SELECT json_agg(role) FROM user_account_roles WHERE user_id = u.id) as roles
FROM users u
LEFT JOIN employees e ON e.id = u.employee_id
LEFT JOIN students s ON s.id = u.student_id
LEFT JOIN parents p ON p.id = u.parent_id
WHERE u.deleted_at IS NULL
ORDER BY u.username ASC;

-- name: CreateUser :one
INSERT INTO users (username, password_hash, display_name, employee_id, student_id, parent_id, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, username, display_name, employee_id, student_id, parent_id, is_active, auth_version, last_login_at, deleted_at, created_at, updated_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: IncrementUserAuthVersion :one
UPDATE users
SET auth_version = auth_version + 1,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING auth_version;

-- name: MarkUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateUserStatus :exec
UPDATE users
SET is_active = $2,
    deleted_at = CASE WHEN $2 = TRUE THEN NULL ELSE deleted_at END,
    auth_version = CASE WHEN is_active = TRUE AND $2 = FALSE THEN auth_version + 1 ELSE auth_version END,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users
SET is_active = FALSE,
    deleted_at = COALESCE(deleted_at, NOW()),
    auth_version = auth_version + 1,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateUserProfileLink :exec
UPDATE users
SET employee_id = $2,
    student_id = $3,
    parent_id = $4,
    display_name = COALESCE(NULLIF(display_name, ''), $5),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetUserRoles :many
SELECT role FROM user_account_roles WHERE user_id = $1;

-- name: AddUserRole :exec
INSERT INTO user_account_roles (user_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: ListUsersByStudentID :many
SELECT id, username, password_hash, display_name, employee_id, created_at, updated_at, student_id, parent_id, is_active, auth_version, last_login_at, deleted_at
FROM users
WHERE student_id = $1
  AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: ListUsersByEmployeeID :many
SELECT id, username, password_hash, display_name, employee_id, created_at, updated_at, student_id, parent_id, is_active, auth_version, last_login_at, deleted_at
FROM users
WHERE employee_id = $1
  AND deleted_at IS NULL
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
        CASE
            WHEN e.tanggal_lahir IS NULL THEN ''::text
            ELSE to_char(e.tanggal_lahir, 'YY')
        END AS birth_year_suffix
    FROM employees e
    WHERE e.is_active = TRUE
), linked_employees AS (
    SELECT
        ae.*,
        linked.id AS existing_user_id
    FROM active_employees ae
    LEFT JOIN users linked ON linked.employee_id = ae.id AND linked.deleted_at IS NULL
), existing_sequences AS (
    SELECT
        substring(u.username FROM length(sqlc.arg(npsn)::text) + 1 FOR 2) AS birth_year_suffix,
        max(substring(u.username FROM length(sqlc.arg(npsn)::text) + 3 FOR 3)::int)::int AS max_sequence
    FROM users u
    WHERE u.username ~ ('^' || sqlc.arg(npsn)::text || '[0-9]{5}$')
    GROUP BY substring(u.username FROM length(sqlc.arg(npsn)::text) + 1 FOR 2)
), sequenced_candidates AS (
    SELECT
        le.id AS employee_id,
        le.nip,
        le.nama,
        le.tanggal_lahir,
        le.birth_year_suffix,
        le.existing_user_id,
        CASE
            WHEN le.tanggal_lahir IS NULL OR le.existing_user_id IS NOT NULL THEN 0
            ELSE (COALESCE(es.max_sequence, 0) + sum(CASE WHEN le.tanggal_lahir IS NOT NULL AND le.existing_user_id IS NULL THEN 1 ELSE 0 END) OVER (
                PARTITION BY le.birth_year_suffix
                ORDER BY le.nama ASC, le.id ASC
            ))::int
        END AS nomor_urut
    FROM linked_employees le
    LEFT JOIN existing_sequences es ON es.birth_year_suffix = le.birth_year_suffix
)
SELECT
    sc.employee_id,
    sc.nip,
    sc.nama,
    sc.tanggal_lahir,
    sc.nomor_urut,
    CASE
        WHEN sc.tanggal_lahir IS NULL OR sc.existing_user_id IS NOT NULL THEN ''::text
        ELSE (sqlc.arg(npsn)::text || sc.birth_year_suffix || lpad(sc.nomor_urut::text, 3, '0'))::text
    END AS generated_username,
    sc.existing_user_id,
    username_user.id AS username_user_id
FROM sequenced_candidates sc
LEFT JOIN users username_user ON username_user.username = (sqlc.arg(npsn)::text || sc.birth_year_suffix || lpad(sc.nomor_urut::text, 3, '0'))
    AND sc.nomor_urut > 0
    AND username_user.deleted_at IS NULL
ORDER BY sc.birth_year_suffix ASC NULLS LAST, sc.nomor_urut ASC, sc.nama ASC, sc.employee_id ASC;
