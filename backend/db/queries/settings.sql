-- name: GetSetting :one
SELECT key, value, updated_at FROM app_settings WHERE key = $1;

-- name: UpsertSetting :exec
INSERT INTO app_settings (key, value, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE
  SET value = EXCLUDED.value, updated_at = NOW();

-- name: ListSettings :many
SELECT key, value, updated_at FROM app_settings ORDER BY key;
