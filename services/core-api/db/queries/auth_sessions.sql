-- name: CreateAuthSession :one
INSERT INTO auth_sessions (id, user_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3, $4)
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
