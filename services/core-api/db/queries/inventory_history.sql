-- name: CreateInventoryItemEvent :one
INSERT INTO inventory_item_events (item_id, actor_user_id, action, summary)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListInventoryItemEventsByItem :many
SELECT
    e.id,
    e.item_id,
    e.actor_user_id,
    e.action,
    e.summary,
    e.created_at,
    COALESCE(u.username, '') AS actor_username
FROM inventory_item_events e
LEFT JOIN users u ON u.id = e.actor_user_id
WHERE e.item_id = $1
ORDER BY e.created_at DESC;
