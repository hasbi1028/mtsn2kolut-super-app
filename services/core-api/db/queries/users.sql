-- name: GetUserByUsername :one
SELECT 
    u.id, u.username, u.password_hash,
    u.display_name,
    u.employee_id, u.student_id, u.parent_id,
    u.is_active, u.auth_version, u.must_change_password, u.password_changed_at, u.last_login_at, u.deleted_at, u.created_at, u.updated_at,
    COALESCE(
      (
        SELECT json_agg(r.code ORDER BY r.code)
        FROM rbac_user_roles ur
        JOIN rbac_roles r ON r.id = ur.role_id
        WHERE ur.user_id = u.id
          AND r.is_active = TRUE
      ),
      (SELECT json_agg(role) FROM user_account_roles WHERE user_id = u.id),
      '[]'::json
    ) as roles
FROM users u
WHERE u.username = $1
  AND u.deleted_at IS NULL;

-- name: GetUserByID :one
SELECT 
    u.id, u.username, u.password_hash,
    u.display_name,
    u.employee_id, u.student_id, u.parent_id,
    u.is_active, u.auth_version, u.must_change_password, u.password_changed_at, u.last_login_at, u.deleted_at, u.created_at, u.updated_at,
    COALESCE(
      (
        SELECT json_agg(r.code ORDER BY r.code)
        FROM rbac_user_roles ur
        JOIN rbac_roles r ON r.id = ur.role_id
        WHERE ur.user_id = u.id
          AND r.is_active = TRUE
      ),
      (SELECT json_agg(role) FROM user_account_roles WHERE user_id = u.id),
      '[]'::json
    ) as roles
FROM users u
WHERE u.id = $1
  AND u.deleted_at IS NULL;

-- name: ListUsers :many
SELECT 
    u.id, u.username,
    COALESCE(NULLIF(u.display_name, ''), e.nama, s.nama, p.nama, u.username) AS display_name,
    u.employee_id, u.student_id, u.parent_id,
    COALESCE(e.nama, s.nama, p.nama, '') AS profile_nama,
    u.is_active, u.must_change_password, u.password_changed_at, u.last_login_at, u.deleted_at, u.created_at,
    COALESCE(
      (
        SELECT json_agg(r.code ORDER BY r.code)
        FROM rbac_user_roles ur
        JOIN rbac_roles r ON r.id = ur.role_id
        WHERE ur.user_id = u.id
          AND r.is_active = TRUE
      ),
      (SELECT json_agg(role) FROM user_account_roles WHERE user_id = u.id),
      '[]'::json
    ) as roles
FROM users u
LEFT JOIN employees e ON e.id = u.employee_id
LEFT JOIN students s ON s.id = u.student_id
LEFT JOIN parents p ON p.id = u.parent_id
WHERE u.deleted_at IS NULL
ORDER BY u.username ASC;

-- name: GetUserAccountSummary :one
SELECT
    u.id,
    u.username,
    COALESCE(NULLIF(u.display_name, ''), e.nama, s.nama, p.nama, u.username)::text AS display_name,
    u.employee_id,
    u.student_id,
    u.parent_id,
    CASE
        WHEN u.employee_id IS NOT NULL THEN 'employee'
        WHEN u.student_id IS NOT NULL THEN 'student'
        WHEN u.parent_id IS NOT NULL THEN 'parent'
        ELSE ''
    END::text AS profile_type,
    COALESCE(e.nama, s.nama, p.nama, '')::text AS profile_nama,
    CASE
        WHEN u.employee_id IS NOT NULL THEN e.phone
        WHEN u.student_id IS NOT NULL THEN s.phone
        WHEN u.parent_id IS NOT NULL THEN p.phone
        ELSE ''
    END::text AS contact_phone,
    CASE
        WHEN u.employee_id IS NOT NULL THEN e.email
        ELSE ''
    END::text AS contact_email,
    CASE
        WHEN u.employee_id IS NOT NULL THEN e.address
        WHEN u.student_id IS NOT NULL THEN s.alamat
        WHEN u.parent_id IS NOT NULL THEN p.address
        ELSE ''
    END::text AS contact_address,
    CASE
        WHEN u.employee_id IS NOT NULL THEN e.photo_url
        WHEN u.student_id IS NOT NULL THEN s.photo_url
        WHEN u.parent_id IS NOT NULL THEN p.photo_url
        ELSE ''
    END::text AS photo_url,
    u.is_active,
    u.must_change_password,
    u.password_changed_at,
    u.last_login_at,
    u.created_at,
    COALESCE(
      (
        SELECT json_agg(r.code ORDER BY r.code)
        FROM rbac_user_roles ur
        JOIN rbac_roles r ON r.id = ur.role_id
        WHERE ur.user_id = u.id
          AND r.is_active = TRUE
      ),
      (SELECT json_agg(role) FROM user_account_roles WHERE user_id = u.id),
      '[]'::json
    ) as roles
FROM users u
LEFT JOIN employees e ON e.id = u.employee_id
LEFT JOIN students s ON s.id = u.student_id
LEFT JOIN parents p ON p.id = u.parent_id
WHERE u.id = $1
  AND u.deleted_at IS NULL;

-- name: UpdateOwnedEmployeeContact :one
UPDATE employees e
SET phone = sqlc.arg(phone),
    email = sqlc.arg(email),
    address = sqlc.arg(address),
    updated_at = NOW()
WHERE e.id = (
    SELECT u.employee_id
    FROM users u
    WHERE u.id = sqlc.arg(user_id)
      AND u.deleted_at IS NULL
      AND u.employee_id IS NOT NULL
)
RETURNING e.phone AS contact_phone, e.email AS contact_email, e.address AS contact_address;

-- name: UpdateOwnedStudentContact :one
UPDATE students s
SET phone = sqlc.arg(phone),
    alamat = sqlc.arg(address),
    updated_at = NOW()
WHERE s.id = (
    SELECT u.student_id
    FROM users u
    WHERE u.id = sqlc.arg(user_id)
      AND u.deleted_at IS NULL
      AND u.student_id IS NOT NULL
)
RETURNING s.phone AS contact_phone, ''::text AS contact_email, s.alamat AS contact_address;

-- name: UpdateOwnedParentContact :one
UPDATE parents p
SET phone = sqlc.arg(phone),
    address = sqlc.arg(address),
    updated_at = NOW()
WHERE p.id = (
    SELECT u.parent_id
    FROM users u
    WHERE u.id = sqlc.arg(user_id)
      AND u.deleted_at IS NULL
      AND u.parent_id IS NOT NULL
)
RETURNING p.phone AS contact_phone, ''::text AS contact_email, p.address AS contact_address;

-- name: UpdateOwnedEmployeeAvatar :one
UPDATE employees e
SET photo_url = sqlc.arg(photo_url),
    updated_at = NOW()
WHERE e.id = (
    SELECT u.employee_id
    FROM users u
    WHERE u.id = sqlc.arg(user_id)
      AND u.deleted_at IS NULL
      AND u.employee_id IS NOT NULL
)
RETURNING e.photo_url;

-- name: UpdateOwnedStudentAvatar :one
UPDATE students s
SET photo_url = sqlc.arg(photo_url),
    updated_at = NOW()
WHERE s.id = (
    SELECT u.student_id
    FROM users u
    WHERE u.id = sqlc.arg(user_id)
      AND u.deleted_at IS NULL
      AND u.student_id IS NOT NULL
)
RETURNING s.photo_url;

-- name: UpdateOwnedParentAvatar :one
UPDATE parents p
SET photo_url = sqlc.arg(photo_url),
    updated_at = NOW()
WHERE p.id = (
    SELECT u.parent_id
    FROM users u
    WHERE u.id = sqlc.arg(user_id)
      AND u.deleted_at IS NULL
      AND u.parent_id IS NOT NULL
)
RETURNING p.photo_url;

-- name: CreateUser :one
INSERT INTO users (username, password_hash, display_name, employee_id, student_id, parent_id, is_active, must_change_password, password_changed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE, NULL)
RETURNING id, username, display_name, employee_id, student_id, parent_id, is_active, auth_version, must_change_password, password_changed_at, last_login_at, deleted_at, created_at, updated_at;

-- name: CreateUserWithMustChangePassword :one
INSERT INTO users (username, password_hash, display_name, employee_id, student_id, parent_id, is_active, must_change_password, password_changed_at)
VALUES (
    sqlc.arg(username),
    sqlc.arg(password_hash),
    sqlc.arg(display_name),
    sqlc.arg(employee_id),
    sqlc.arg(student_id),
    sqlc.arg(parent_id),
    sqlc.arg(is_active),
    sqlc.arg(must_change_password),
    CASE WHEN sqlc.arg(must_change_password)::boolean THEN NULL ELSE NOW() END
)
RETURNING id, username, display_name, employee_id, student_id, parent_id, is_active, auth_version, must_change_password, password_changed_at, last_login_at, deleted_at, created_at, updated_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ChangeUserPasswordAndInvalidate :one
WITH updated AS (
  UPDATE users
  SET password_hash = sqlc.arg(password_hash),
      must_change_password = FALSE,
      password_changed_at = NOW(),
      auth_version = auth_version + 1,
      updated_at = NOW()
  WHERE users.id = sqlc.arg(id)
    AND users.deleted_at IS NULL
  RETURNING users.id, users.auth_version
), revoked AS (
  UPDATE auth_sessions
  SET revoked_at = NOW(),
      updated_at = NOW()
  WHERE user_id = (SELECT updated.id FROM updated)
    AND revoked_at IS NULL
  RETURNING auth_sessions.id
)
SELECT updated.auth_version FROM updated;

-- name: MarkUserPasswordChanged :exec
UPDATE users
SET must_change_password = FALSE,
    password_changed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: MarkUserMustChangePassword :exec
UPDATE users
SET must_change_password = TRUE,
    password_changed_at = NULL,
    updated_at = NOW()
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

-- name: SyncLegacyUserRolesFromRbac :exec
WITH legacy_roles AS (
  SELECT r.code::user_role AS role
  FROM rbac_user_roles ur
  JOIN rbac_roles r ON r.id = ur.role_id
  WHERE ur.user_id = sqlc.arg(user_id)
    AND r.code IN ('admin', 'guru', 'staf', 'kesiswaan', 'siswa', 'ortu')
)
INSERT INTO user_account_roles (user_id, role)
SELECT sqlc.arg(user_id), role
FROM legacy_roles
ON CONFLICT DO NOTHING;

-- name: ListUsersByStudentID :many
SELECT id, username, password_hash, display_name, employee_id, created_at, updated_at, student_id, parent_id, is_active, auth_version, must_change_password, password_changed_at, last_login_at, deleted_at
FROM users
WHERE student_id = $1
  AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: ListUsersByEmployeeID :many
SELECT id, username, password_hash, display_name, employee_id, created_at, updated_at, student_id, parent_id, is_active, auth_version, must_change_password, password_changed_at, last_login_at, deleted_at
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

-- name: ListEmployeeProfileCandidates :many
SELECT
    e.id,
    e.nama,
    COALESCE(NULLIF(e.nip, ''), e.pegawai_uid, '')::text AS identifier,
    linked.id AS linked_user_id,
    COALESCE(linked.username, '')::text AS linked_username
FROM employees e
LEFT JOIN LATERAL (
    SELECT u.id, u.username
    FROM users u
    WHERE u.employee_id = e.id
      AND u.deleted_at IS NULL
    ORDER BY u.created_at DESC, u.id DESC
    LIMIT 1
) linked ON TRUE
WHERE e.is_active = TRUE
  AND (sqlc.arg(include_linked)::boolean OR linked.id IS NULL)
  AND (
    btrim(sqlc.arg(q)::text) = ''
    OR e.nama ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
    OR COALESCE(e.nip, '') ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
    OR COALESCE(e.pegawai_uid, '') ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
  )
ORDER BY e.nama ASC, e.id ASC
LIMIT sqlc.arg(limit_count)::int;

-- name: ListStudentProfileCandidatesByClass :many
SELECT
    s.id,
    s.nama,
    COALESCE(NULLIF(s.nisn, ''), s.nis, '')::text AS identifier,
    s.class_id AS class_id,
    COALESCE(NULLIF(c.name, ''), c.code, '')::text AS class_name,
    linked.id AS linked_user_id,
    COALESCE(linked.username, '')::text AS linked_username
FROM students s
JOIN school_classes c
    ON c.id = s.class_id
LEFT JOIN LATERAL (
    SELECT u.id, u.username
    FROM users u
    WHERE u.student_id = s.id
      AND u.deleted_at IS NULL
    ORDER BY u.created_at DESC, u.id DESC
    LIMIT 1
) linked ON TRUE
WHERE s.is_active = TRUE
  AND s.status = 'active'
  AND s.class_id = sqlc.arg(class_id)
  AND (sqlc.arg(include_linked)::boolean OR linked.id IS NULL)
  AND (
    btrim(sqlc.arg(q)::text) = ''
    OR s.nama ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
    OR s.nis ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
    OR COALESCE(s.nisn, '') ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
  )
ORDER BY s.nama ASC, s.id ASC
LIMIT sqlc.arg(limit_count)::int;

-- name: ListParentProfileCandidatesByChildClass :many
WITH matching_parents AS (
    SELECT
        p.id,
        p.nama,
        COALESCE(p.phone, '')::text AS identifier,
        linked.id AS linked_user_id,
        COALESCE(linked.username, '')::text AS linked_username
    FROM parents p
    JOIN parent_students ps
        ON ps.parent_id = p.id
    JOIN students s
        ON s.id = ps.student_id
    LEFT JOIN LATERAL (
        SELECT u.id, u.username
        FROM users u
        WHERE u.parent_id = p.id
          AND u.deleted_at IS NULL
        ORDER BY u.created_at DESC, u.id DESC
        LIMIT 1
    ) linked ON TRUE
    WHERE s.is_active = TRUE
      AND s.status = 'active'
      AND s.class_id = sqlc.arg(class_id)
      AND (sqlc.arg(include_linked)::boolean OR linked.id IS NULL)
      AND (
        btrim(sqlc.arg(q)::text) = ''
        OR p.nama ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
        OR COALESCE(p.phone, '') ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
        OR s.nama ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
        OR s.nis ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
        OR COALESCE(s.nisn, '') ILIKE ('%' || btrim(sqlc.arg(q)::text) || '%')
      )
    GROUP BY p.id, p.nama, p.phone, linked.id, linked.username
    ORDER BY p.nama ASC, p.id ASC
    LIMIT sqlc.arg(limit_count)::int
)
SELECT
    mp.id,
    mp.nama,
    mp.identifier,
    s.class_id AS class_id,
    COALESCE(NULLIF(c.name, ''), c.code, '')::text AS class_name,
    mp.linked_user_id,
    mp.linked_username,
    s.id AS child_id,
    s.nama AS child_nama
FROM matching_parents mp
JOIN parent_students ps
    ON ps.parent_id = mp.id
JOIN students s
    ON s.id = ps.student_id
JOIN school_classes c
    ON c.id = s.class_id
WHERE s.is_active = TRUE
  AND s.status = 'active'
  AND s.class_id = sqlc.arg(class_id)
ORDER BY mp.nama ASC, mp.id ASC, ps.is_primary_contact DESC, s.nama ASC, s.id ASC;

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

-- name: ListOwnAccountChangeHistory :many
WITH relevant_audit AS (
    SELECT a.*
    FROM audit_logs a
    LEFT JOIN profile_change_requests pcr
        ON pcr.id::text = COALESCE(NULLIF(a.metadata->>'request_id', ''), a.entity_id)
    WHERE a.action IN (
        'AUTH_ACCOUNT_CONTACT_UPDATE',
        'AUTH_ACCOUNT_AVATAR_UPDATE',
        'AUTH_ACCOUNT_AVATAR_DELETE',
        'ACCOUNT_CHANGE_REQUEST_CREATED',
        'ACCOUNT_CHANGE_REQUEST_CANCELLED',
        'ACCOUNT_CHANGE_REQUEST_APPROVED',
        'ACCOUNT_CHANGE_REQUEST_REJECTED'
    )
      AND (
        (
            a.action IN (
                'AUTH_ACCOUNT_CONTACT_UPDATE',
                'AUTH_ACCOUNT_AVATAR_UPDATE',
                'AUTH_ACCOUNT_AVATAR_DELETE'
            )
            AND a.user_id = sqlc.arg(user_id)
        )
        OR (
            a.action IN (
                'ACCOUNT_CHANGE_REQUEST_CREATED',
                'ACCOUNT_CHANGE_REQUEST_CANCELLED',
                'ACCOUNT_CHANGE_REQUEST_APPROVED',
                'ACCOUNT_CHANGE_REQUEST_REJECTED'
            )
            AND (
                a.metadata->>'requester_user_id' = sqlc.arg(user_id)::text
                OR pcr.requester_user_id = sqlc.arg(user_id)
            )
        )
      )
)
SELECT
    a.action,
    CASE
        WHEN a.action = 'AUTH_ACCOUNT_CONTACT_UPDATE' THEN 'contact'
        WHEN a.action IN ('AUTH_ACCOUNT_AVATAR_UPDATE', 'AUTH_ACCOUNT_AVATAR_DELETE') THEN 'avatar'
        ELSE COALESCE(NULLIF(a.metadata->>'field_key', ''), pcr.field_key, '')
    END::text AS field_key,
    CASE
        WHEN a.action = 'ACCOUNT_CHANGE_REQUEST_CREATED' THEN 'pending'
        WHEN a.action = 'ACCOUNT_CHANGE_REQUEST_CANCELLED' THEN 'cancelled'
        WHEN a.action = 'ACCOUNT_CHANGE_REQUEST_APPROVED' THEN 'approved'
        WHEN a.action = 'ACCOUNT_CHANGE_REQUEST_REJECTED' THEN 'rejected'
        ELSE 'completed'
    END::text AS status,
    a.created_at,
    CASE
        WHEN a.action IN ('ACCOUNT_CHANGE_REQUEST_APPROVED', 'ACCOUNT_CHANGE_REQUEST_REJECTED') THEN COALESCE(reviewer.username, '')
        ELSE ''
    END::text AS reviewer_username,
    CASE
        WHEN a.action IN ('ACCOUNT_CHANGE_REQUEST_APPROVED', 'ACCOUNT_CHANGE_REQUEST_REJECTED') THEN COALESCE(pcr.review_note, '')
        ELSE ''
    END::text AS review_note
FROM relevant_audit a
LEFT JOIN profile_change_requests pcr
    ON pcr.id::text = COALESCE(NULLIF(a.metadata->>'request_id', ''), a.entity_id)
LEFT JOIN users reviewer
    ON reviewer.id = a.user_id
   AND a.action IN ('ACCOUNT_CHANGE_REQUEST_APPROVED', 'ACCOUNT_CHANGE_REQUEST_REJECTED')
ORDER BY a.created_at DESC
LIMIT sqlc.arg(limit_count)::int;

-- name: DeleteOldAuditLogs :execrows
DELETE FROM audit_logs
WHERE created_at < NOW() - INTERVAL '90 days';
-- name: ListEmployeeAccountGenerationCandidates :many
WITH active_employees AS (
    SELECT
        e.id,
        COALESCE(e.nip, '')::text AS nip,
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

-- name: ListStudentAccountGenerationCandidates :many
WITH active_students AS (
    SELECT
        s.id AS student_id,
        COALESCE(s.nis, '')::text AS nis,
        COALESCE(s.nisn, '')::text AS nisn,
        s.nama,
        CASE
            WHEN btrim(COALESCE(s.nisn, '')) <> '' THEN btrim(s.nisn)
            WHEN btrim(COALESCE(s.nis, '')) <> '' THEN ('s' || btrim(s.nis))::text
            ELSE ''::text
        END AS base_username
    FROM students s
    WHERE s.is_active = TRUE
), linked_students AS (
    SELECT
        active_students.*,
        linked.id AS existing_user_id,
        COALESCE(linked.username, '')::text AS existing_username
    FROM active_students
    LEFT JOIN users linked
        ON linked.student_id = active_students.student_id
       AND linked.deleted_at IS NULL
)
SELECT
    linked_students.student_id,
    linked_students.nis,
    linked_students.nisn,
    linked_students.nama,
    linked_students.base_username,
    linked_students.existing_user_id,
    linked_students.existing_username,
    username_user.id AS username_user_id,
    ARRAY(
        SELECT u.username
        FROM users u
        WHERE linked_students.base_username <> ''
          AND u.deleted_at IS NULL
          AND (
            u.username = linked_students.base_username
            OR u.username LIKE linked_students.base_username || '-__'
          )
        ORDER BY u.username ASC
    )::text[] AS username_collisions
FROM linked_students
LEFT JOIN users username_user
    ON username_user.username = linked_students.base_username
   AND linked_students.base_username <> ''
   AND username_user.deleted_at IS NULL
ORDER BY linked_students.nama ASC, linked_students.student_id ASC;

-- name: ListParentAccountGenerationCandidates :many
WITH child_counts AS (
    SELECT
        ps.parent_id,
        count(*)::int AS child_count
    FROM parent_students ps
    GROUP BY ps.parent_id
), basis_children AS (
    SELECT DISTINCT ON (ps.parent_id)
        ps.parent_id,
        s.id AS basis_student_id,
        COALESCE(s.nisn, '')::text AS basis_student_nisn
    FROM parent_students ps
    JOIN students s ON s.id = ps.student_id
    ORDER BY
        ps.parent_id,
        ps.is_primary_contact DESC,
        CASE WHEN btrim(COALESCE(s.nisn, '')) <> '' THEN 0 ELSE 1 END,
        s.nama ASC,
        s.id ASC
), parent_candidates AS (
    SELECT
        p.id AS parent_id,
        p.nama,
        COALESCE(p.phone, '')::text AS phone,
        COALESCE(child_counts.child_count, 0)::int AS child_count,
        basis_children.basis_student_id,
        COALESCE(basis_children.basis_student_nisn, '')::text AS basis_student_nisn,
        regexp_replace(COALESCE(p.phone, ''), '\D', '', 'g')::text AS phone_digits
    FROM parents p
    LEFT JOIN child_counts ON child_counts.parent_id = p.id
    LEFT JOIN basis_children ON basis_children.parent_id = p.id
), linked_parents AS (
    SELECT
        parent_candidates.*,
        CASE
            WHEN btrim(parent_candidates.basis_student_nisn) <> '' THEN ('ortu' || btrim(parent_candidates.basis_student_nisn))::text
            WHEN parent_candidates.phone_digits <> '' THEN ('ortu' || right(parent_candidates.phone_digits, 8))::text
            ELSE ''::text
        END AS base_username,
        linked.id AS existing_user_id,
        COALESCE(linked.username, '')::text AS existing_username
    FROM parent_candidates
    LEFT JOIN users linked
        ON linked.parent_id = parent_candidates.parent_id
       AND linked.deleted_at IS NULL
)
SELECT
    linked_parents.parent_id,
    linked_parents.nama,
    linked_parents.phone,
    linked_parents.child_count,
    linked_parents.basis_student_id,
    linked_parents.basis_student_nisn,
    linked_parents.base_username,
    linked_parents.existing_user_id,
    linked_parents.existing_username,
    username_user.id AS username_user_id,
    ARRAY(
        SELECT u.username
        FROM users u
        WHERE linked_parents.base_username <> ''
          AND u.deleted_at IS NULL
          AND (
            u.username = linked_parents.base_username
            OR u.username LIKE linked_parents.base_username || '-__'
          )
        ORDER BY u.username ASC
    )::text[] AS username_collisions
FROM linked_parents
LEFT JOIN users username_user
    ON username_user.username = linked_parents.base_username
   AND linked_parents.base_username <> ''
   AND username_user.deleted_at IS NULL
ORDER BY linked_parents.nama ASC, linked_parents.parent_id ASC;
