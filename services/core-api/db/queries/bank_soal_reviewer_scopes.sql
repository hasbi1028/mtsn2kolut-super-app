-- name: ListBankSoalReviewerScopes :many
SELECT
    rs.id,
    rs.user_id,
    u.username,
    COALESCE(NULLIF(btrim(u.display_name), ''), NULLIF(btrim(e.nama), ''), u.username)::text AS user_display_name,
    COALESCE(user_roles.roles, '[]'::json)::json AS user_roles,
    rs.subject_id,
    COALESCE(s.name, '')::text AS subject_name,
    COALESCE(s.code, '')::text AS subject_code,
    rs.grade_level,
    rs.can_review,
    rs.can_approve,
    rs.assigned_by,
    COALESCE(NULLIF(btrim(assigned_user.display_name), ''), NULLIF(btrim(assigned_emp.nama), ''), assigned_user.username, '')::text AS assigned_by_display_name,
    rs.created_at,
    rs.updated_at
FROM bank_soal_reviewer_scopes rs
JOIN users u ON u.id = rs.user_id
LEFT JOIN employees e ON e.id = u.employee_id
LEFT JOIN subjects s ON s.id = rs.subject_id
LEFT JOIN users assigned_user ON assigned_user.id = rs.assigned_by
LEFT JOIN employees assigned_emp ON assigned_emp.id = assigned_user.employee_id
LEFT JOIN LATERAL (
    SELECT json_agg(r.code ORDER BY r.code) AS roles
    FROM rbac_user_roles ur
    JOIN rbac_roles r ON r.id = ur.role_id
    WHERE ur.user_id = u.id
      AND r.is_active = TRUE
) user_roles ON TRUE
WHERE u.deleted_at IS NULL
ORDER BY user_display_name ASC, subject_name ASC, rs.grade_level ASC NULLS FIRST;

-- name: UpsertBankSoalReviewerScope :one
WITH updated AS (
    UPDATE bank_soal_reviewer_scopes AS rs
    SET can_review = sqlc.arg(can_review),
        can_approve = sqlc.arg(can_approve),
        assigned_by = sqlc.arg(assigned_by),
        updated_at = NOW()
    WHERE rs.user_id = sqlc.arg(user_id)
      AND (
          (rs.subject_id IS NULL AND sqlc.narg(subject_id)::uuid IS NULL)
          OR rs.subject_id = sqlc.narg(subject_id)::uuid
      )
      AND (
          (rs.grade_level IS NULL AND sqlc.narg(grade_level)::smallint IS NULL)
          OR rs.grade_level = sqlc.narg(grade_level)::smallint
      )
    RETURNING *
),
inserted AS (
    INSERT INTO bank_soal_reviewer_scopes (
        user_id,
        subject_id,
        grade_level,
        can_review,
        can_approve,
        assigned_by
    )
    SELECT
        sqlc.arg(user_id),
        sqlc.narg(subject_id),
        sqlc.narg(grade_level),
        sqlc.arg(can_review),
        sqlc.arg(can_approve),
        sqlc.arg(assigned_by)
    WHERE NOT EXISTS (SELECT 1 FROM updated)
    ON CONFLICT DO NOTHING
    RETURNING *
)
SELECT * FROM updated
UNION ALL
SELECT * FROM inserted
LIMIT 1;

-- name: DeleteBankSoalReviewerScope :one
DELETE FROM bank_soal_reviewer_scopes
WHERE id = $1
RETURNING id;

-- name: CanBankSoalUserReview :one
SELECT EXISTS (
    SELECT 1
    FROM bank_soal_reviewer_scopes rs
    WHERE rs.user_id = sqlc.arg(user_id)
      AND rs.can_review = TRUE
      AND (
          rs.subject_id IS NULL
          OR (sqlc.narg(subject_id)::uuid IS NOT NULL AND rs.subject_id = sqlc.narg(subject_id)::uuid)
      )
      AND (
          rs.grade_level IS NULL
          OR (sqlc.narg(grade_level)::smallint IS NOT NULL AND rs.grade_level = sqlc.narg(grade_level)::smallint)
      )
)::bool;

-- name: CanBankSoalUserApprove :one
SELECT EXISTS (
    SELECT 1
    FROM bank_soal_reviewer_scopes rs
    WHERE rs.user_id = sqlc.arg(user_id)
      AND rs.can_approve = TRUE
      AND (
          rs.subject_id IS NULL
          OR (sqlc.narg(subject_id)::uuid IS NOT NULL AND rs.subject_id = sqlc.narg(subject_id)::uuid)
      )
      AND (
          rs.grade_level IS NULL
          OR (sqlc.narg(grade_level)::smallint IS NOT NULL AND rs.grade_level = sqlc.narg(grade_level)::smallint)
      )
)::bool;
