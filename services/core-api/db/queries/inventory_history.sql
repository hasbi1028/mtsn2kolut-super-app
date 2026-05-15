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
    COALESCE(u.username, '') AS actor_username,
    COALESCE(
        NULLIF(btrim(eu.nama), ''),
        NULLIF(btrim(s.nama), ''),
        NULLIF(btrim(p.nama), ''),
        NULLIF(btrim(u.display_name), ''),
        u.username,
        ''
    )::text AS actor_display_name
FROM inventory_item_events e
LEFT JOIN users u ON u.id = e.actor_user_id
LEFT JOIN employees eu ON eu.id = u.employee_id
LEFT JOIN students s ON s.id = u.student_id
LEFT JOIN parents p ON p.id = u.parent_id
WHERE e.item_id = $1
ORDER BY e.created_at DESC;
