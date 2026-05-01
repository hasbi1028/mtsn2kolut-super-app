-- name: GetUserUIPreferences :one
SELECT user_id, sidebar_pinned, sidebar_recent, created_at, updated_at
FROM user_ui_preferences
WHERE user_id = $1;

-- name: UpsertUserUIPreferences :one
INSERT INTO user_ui_preferences (user_id, sidebar_pinned, sidebar_recent)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO UPDATE
SET sidebar_pinned = EXCLUDED.sidebar_pinned,
    sidebar_recent = EXCLUDED.sidebar_recent,
    updated_at = NOW()
RETURNING user_id, sidebar_pinned, sidebar_recent, created_at, updated_at;
