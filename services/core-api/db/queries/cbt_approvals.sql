-- name: ListCbtApprovalRecords :many
SELECT
  ar.id,
  ar.entity_type,
  ar.entity_id,
  ar.approval_type,
  ar.status,
  ar.approved_by,
  COALESCE(NULLIF(approved_user.display_name, ''), approved_employee.nama, approved_user.username, '')::text AS approved_by_display_name,
  COALESCE(approved_user.username, '')::text AS approved_by_username,
  ar.approved_at,
  ar.revoked_by,
  COALESCE(NULLIF(revoked_user.display_name, ''), revoked_employee.nama, revoked_user.username, '')::text AS revoked_by_display_name,
  COALESCE(revoked_user.username, '')::text AS revoked_by_username,
  ar.revoked_at,
  ar.notes,
  ar.created_at,
  ar.updated_at
FROM cbt_approval_records ar
LEFT JOIN users approved_user ON approved_user.id = ar.approved_by
LEFT JOIN employees approved_employee ON approved_employee.id = approved_user.employee_id
LEFT JOIN users revoked_user ON revoked_user.id = ar.revoked_by
LEFT JOIN employees revoked_employee ON revoked_employee.id = revoked_user.employee_id
WHERE (sqlc.narg(entity_type)::text IS NULL OR ar.entity_type = sqlc.narg(entity_type)::text)
  AND (sqlc.narg(entity_id)::uuid IS NULL OR ar.entity_id = sqlc.narg(entity_id)::uuid)
ORDER BY ar.approved_at DESC NULLS LAST, ar.created_at DESC;

-- name: CreateCbtApprovalRecord :one
INSERT INTO cbt_approval_records (
  entity_type,
  entity_id,
  approval_type,
  status,
  approved_by,
  approved_at,
  notes
) VALUES (
  sqlc.arg(entity_type),
  sqlc.arg(entity_id),
  sqlc.arg(approval_type),
  'approved',
  sqlc.arg(approved_by),
  NOW(),
  sqlc.arg(notes)
)
ON CONFLICT (entity_type, entity_id, approval_type) WHERE status = 'approved'
DO UPDATE SET
  approved_by = EXCLUDED.approved_by,
  approved_at = EXCLUDED.approved_at,
  notes = EXCLUDED.notes,
  updated_at = NOW()
RETURNING *;

-- name: RevokeCbtApprovalRecord :one
UPDATE cbt_approval_records
SET status = 'revoked',
    revoked_by = sqlc.arg(revoked_by),
    revoked_at = NOW(),
    notes = CASE
      WHEN TRIM(sqlc.arg(notes)::text) = '' THEN notes
      WHEN notes = '' THEN 'Revoke: ' || TRIM(sqlc.arg(notes)::text)
      ELSE notes || E'\nRevoke: ' || TRIM(sqlc.arg(notes)::text)
    END,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'approved'
RETURNING *;
