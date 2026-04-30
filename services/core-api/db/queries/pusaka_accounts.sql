-- name: UpsertPusakaAccount :one
INSERT INTO pusaka_accounts (employee_id, pusaka_username, pusaka_password, is_enabled)
VALUES ($1, $2, $3, $4)
ON CONFLICT (employee_id) DO UPDATE
SET pusaka_username = EXCLUDED.pusaka_username,
    pusaka_password = CASE
      WHEN EXCLUDED.pusaka_password = '' THEN pusaka_accounts.pusaka_password
      ELSE EXCLUDED.pusaka_password
    END,
    is_enabled      = EXCLUDED.is_enabled,
    updated_at      = NOW()
RETURNING *;

-- name: GetPusakaAccountByEmployeeID :one
SELECT id, employee_id, pusaka_username, pusaka_password, is_enabled, created_at, updated_at
FROM pusaka_accounts
WHERE employee_id = $1;

-- name: DeletePusakaAccountByEmployeeID :exec
DELETE FROM pusaka_accounts
WHERE employee_id = $1;
