-- name: CreateAuthSession :one
INSERT INTO auth_sessions (id, user_id, refresh_token_hash, expires_at, last_used_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING *;

-- name: GetAuthSession :one
SELECT *
FROM auth_sessions
WHERE id = $1;

-- name: RevokeAuthSession :exec
UPDATE auth_sessions
SET revoked_at = NOW(), updated_at = NOW()
WHERE id = $1
  AND revoked_at IS NULL;

-- name: RevokeAllAuthSessionsForUser :execrows
UPDATE auth_sessions
SET revoked_at = NOW(), updated_at = NOW()
WHERE user_id = $1
  AND revoked_at IS NULL;

-- name: ListActiveAuthSessionsByUser :many
SELECT *
FROM auth_sessions
WHERE user_id = $1
  AND revoked_at IS NULL
  AND expires_at > NOW()
ORDER BY last_used_at DESC, created_at DESC;

-- name: RevokeOwnedAuthSession :execrows
UPDATE auth_sessions
SET revoked_at = NOW(), updated_at = NOW()
WHERE user_id = $1
  AND id = $2
  AND revoked_at IS NULL;
