-- name: ListAppNotifications :many
SELECT
    id,
    user_id,
    category,
    title,
    body,
    entity_type,
    entity_id,
    link_path,
    dedupe_key,
    read_at,
    created_at
FROM app_notifications
WHERE user_id = sqlc.arg(user_id)
  AND (sqlc.arg(unread_only)::BOOLEAN = FALSE OR read_at IS NULL)
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count)::INT
OFFSET sqlc.arg(offset_count)::INT;

-- name: CountUnreadAppNotifications :one
SELECT COUNT(*)::BIGINT
FROM app_notifications
WHERE user_id = sqlc.arg(user_id)
  AND read_at IS NULL;

-- name: MarkAppNotificationRead :one
UPDATE app_notifications
SET read_at = COALESCE(read_at, NOW())
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, category, title, body, entity_type, entity_id, link_path, dedupe_key, read_at, created_at;

-- name: MarkAllAppNotificationsRead :execrows
UPDATE app_notifications
SET read_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND read_at IS NULL;

-- name: CreateDocumentCycleReminderNotifications :execrows
INSERT INTO app_notifications (
    user_id,
    category,
    title,
    body,
    entity_type,
    entity_id,
    link_path,
    dedupe_key
)
SELECT
    u.id,
    CASE
        WHEN o.due_date < sqlc.arg(today)::DATE THEN 'document_cycle_late'
        ELSE 'document_cycle_reminder'
    END,
    CASE
        WHEN o.due_date < sqlc.arg(today)::DATE THEN 'Dokumen melewati tenggat'
        ELSE 'Dokumen mendekati tenggat'
    END,
    CASE
        WHEN o.due_date < sqlc.arg(today)::DATE THEN
            FORMAT(
                '%s melewati tenggat pada %s. PIC: %s.',
                c.title,
                TO_CHAR(o.due_date, 'DD Mon YYYY'),
                COALESCE(NULLIF(re.nama, ''), 'penyusun')
            )
        ELSE
            FORMAT(
                '%s perlu disiapkan sebelum %s. PIC: %s.',
                c.title,
                TO_CHAR(o.due_date, 'DD Mon YYYY'),
                COALESCE(NULLIF(re.nama, ''), 'penyusun')
            )
    END,
    'document_cycle_obligation',
    o.id::TEXT,
    FORMAT('/document-cycles?selected_obligation=%s&period_year=%s', o.id, o.period_year),
    FORMAT('document-cycle-reminder:%s:%s', o.id, TO_CHAR(sqlc.arg(today)::DATE, 'YYYY-MM-DD'))
FROM document_cycle_obligations o
JOIN document_cycle_catalogs c ON c.id = o.catalog_id
JOIN users u ON u.employee_id = o.responsible_employee_id
LEFT JOIN employees re ON re.id = o.responsible_employee_id
WHERE o.status <> 'completed'
  AND o.responsible_employee_id IS NOT NULL
  AND u.is_active = TRUE
  AND o.reminder_date <= sqlc.arg(today)::DATE
ON CONFLICT (user_id, dedupe_key) DO NOTHING;
