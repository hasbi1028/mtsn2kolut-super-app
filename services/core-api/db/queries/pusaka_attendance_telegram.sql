-- name: GetPusakaAttendanceTelegramSettings :one
SELECT id, settings_key, is_enabled, send_time, timezone, target_chat_id,
       include_caption, include_image, report_mode, created_at, updated_at
FROM pusaka_attendance_telegram_settings
WHERE settings_key = 'default';

-- name: UpsertPusakaAttendanceTelegramSettings :one
INSERT INTO pusaka_attendance_telegram_settings (
  settings_key, is_enabled, send_time, timezone, target_chat_id,
  include_caption, include_image, report_mode
)
VALUES ('default', $1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (settings_key) DO UPDATE
SET is_enabled = EXCLUDED.is_enabled,
    send_time = EXCLUDED.send_time,
    timezone = EXCLUDED.timezone,
    target_chat_id = EXCLUDED.target_chat_id,
    include_caption = EXCLUDED.include_caption,
    include_image = EXCLUDED.include_image,
    report_mode = EXCLUDED.report_mode,
    updated_at = now()
RETURNING id, settings_key, is_enabled, send_time, timezone, target_chat_id,
          include_caption, include_image, report_mode, created_at, updated_at;

-- name: ListPusakaAttendanceTelegramReportRows :many
SELECT e.id AS employee_id,
       e.nama AS employee_nama,
       COALESCE(e.nip, '')::text AS employee_nip,
       COALESCE(ar.jam_masuk, '')::text AS jam_masuk,
       COALESCE(ar.jam_pulang, '')::text AS jam_pulang
FROM employees e
LEFT JOIN attendance_records ar
  ON ar.employee_id = e.id
 AND ar.tanggal = $1
WHERE e.is_active = TRUE
  AND e.employment_type IN ('pns', 'pppk')
ORDER BY e.nama ASC;

-- name: HasPusakaAttendanceTelegramScheduledLog :one
SELECT EXISTS (
  SELECT 1
  FROM pusaka_attendance_telegram_logs
  WHERE report_date = $1
    AND target_chat_id = $2
    AND send_mode = 'scheduled'
)::boolean;

-- name: CreatePusakaAttendanceTelegramLog :one
INSERT INTO pusaka_attendance_telegram_logs (
  report_date, target_chat_id, send_mode, status,
  telegram_message_id, error_message, requested_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, report_date, target_chat_id, send_mode, status,
          telegram_message_id, error_message, requested_by, sent_at;

-- name: ListPusakaAttendanceTelegramLogs :many
SELECT id,
       report_date,
       CASE
         WHEN length(target_chat_id) <= 4 THEN repeat('*', length(target_chat_id))
         ELSE repeat('*', GREATEST(length(target_chat_id) - 4, 0)) || right(target_chat_id, 4)
       END::text AS target_chat_id_masked,
       send_mode,
       status,
       telegram_message_id,
       error_message,
       requested_by,
       sent_at
FROM pusaka_attendance_telegram_logs
ORDER BY sent_at DESC
LIMIT $1 OFFSET $2;
