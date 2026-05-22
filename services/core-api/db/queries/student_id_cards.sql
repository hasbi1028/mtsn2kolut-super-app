-- name: CreateStudentIDCard :one
INSERT INTO student_id_cards (student_id, card_no, token_hash, token_hint, status, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, 'active', sqlc.narg(created_by_user_id), sqlc.narg(updated_by_user_id))
RETURNING *;

-- name: ReissueStudentIDCard :one
WITH old_card AS (
  SELECT student_id_cards.*
  FROM student_id_cards
  WHERE student_id_cards.id = sqlc.arg(old_id)
    AND student_id_cards.status = 'active'
  FOR UPDATE
), replaced AS (
  UPDATE student_id_cards c
  SET status = 'replaced',
      revoked_at = COALESCE(c.revoked_at, NOW()),
      revoked_reason = COALESCE(NULLIF(sqlc.arg(reason)::TEXT, ''), c.revoked_reason),
      updated_by_user_id = sqlc.narg(actor_id),
      updated_at = NOW()
  FROM old_card o
  WHERE c.id = o.id
  RETURNING c.*
), new_card AS (
  INSERT INTO student_id_cards (student_id, card_no, token_hash, token_hint, status, reissued_from_id, created_by_user_id, updated_by_user_id)
  SELECT r.student_id, sqlc.arg(card_no), sqlc.arg(token_hash), sqlc.arg(token_hint), 'active', r.id, sqlc.narg(actor_id), sqlc.narg(actor_id)
  FROM replaced r
  RETURNING *
)
SELECT n.*, s.nis, s.nisn, s.nama, s.photo_url, s.is_active AS student_is_active,
       COALESCE(sc.name, '') AS class_name, COALESCE(sc.code, '') AS class_code
FROM new_card n
JOIN students s ON s.id = n.student_id
LEFT JOIN school_classes sc ON sc.id = s.class_id;

-- name: GetStudentIDCard :one
SELECT c.*, s.nis, s.nisn, s.nama, s.photo_url, s.is_active AS student_is_active,
       COALESCE(sc.name, '') AS class_name, COALESCE(sc.code, '') AS class_code
FROM student_id_cards c
JOIN students s ON s.id = c.student_id
LEFT JOIN school_classes sc ON sc.id = s.class_id
WHERE c.id = $1;

-- name: GetStudentIDCardByHash :one
SELECT c.*, s.nis, s.nisn, s.nama, s.photo_url, s.is_active AS student_is_active,
       COALESCE(sc.name, '') AS class_name, COALESCE(sc.code, '') AS class_code
FROM student_id_cards c
JOIN students s ON s.id = c.student_id
LEFT JOIN school_classes sc ON sc.id = s.class_id
WHERE c.token_hash = $1;

-- name: GetActiveStudentIDCardByStudent :one
SELECT * FROM student_id_cards
WHERE student_id = $1 AND status = 'active'
ORDER BY created_at DESC
LIMIT 1;

-- name: ListStudentIDCards :many
SELECT c.*, s.nis, s.nisn, s.nama, s.photo_url, s.is_active AS student_is_active,
       COALESCE(sc.name, '') AS class_name, COALESCE(sc.code, '') AS class_code
FROM student_id_cards c
JOIN students s ON s.id = c.student_id
LEFT JOIN school_classes sc ON sc.id = s.class_id
WHERE (sqlc.arg(search)::TEXT = '' OR s.nama ILIKE '%' || sqlc.arg(search) || '%' OR s.nis ILIKE '%' || sqlc.arg(search) || '%' OR c.card_no ILIKE '%' || sqlc.arg(search) || '%')
  AND (sqlc.arg(status)::TEXT = '' OR c.status = sqlc.arg(status))
ORDER BY c.created_at DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: UpdateStudentIDCardStatus :one
UPDATE student_id_cards
SET status = $2,
    revoked_at = CASE WHEN $2 IN ('lost','revoked','replaced','expired','suspended') THEN COALESCE(revoked_at, NOW()) ELSE NULL END,
    revoked_reason = COALESCE(sqlc.narg(revoked_reason), revoked_reason),
    updated_by_user_id = sqlc.narg(updated_by_user_id),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkStudentIDCardPrinted :one
UPDATE student_id_cards
SET printed_at = COALESCE(printed_at, NOW()), updated_by_user_id = sqlc.narg(updated_by_user_id), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: InsertStudentIDCardEvent :one
INSERT INTO student_id_card_events (card_id, student_id, event_type, actor_user_id, source, metadata)
VALUES ($1, $2, $3, sqlc.narg(actor_user_id), $4, $5)
RETURNING *;

-- name: ListStudentIDCardEvents :many
SELECT * FROM student_id_card_events
WHERE card_id = $1
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: InsertStudentIDCardAuditLog :one
INSERT INTO student_id_card_audit_logs (card_id, student_id, actor_user_id, action, target, metadata)
VALUES (sqlc.narg(card_id), sqlc.narg(student_id), sqlc.narg(actor_user_id), $1, $2, $3)
RETURNING *;

-- name: ListStudentIDCardAuditLogs :many
SELECT * FROM student_id_card_audit_logs
WHERE (sqlc.narg(card_id)::UUID IS NULL OR card_id = sqlc.narg(card_id))
  AND (sqlc.narg(student_id)::UUID IS NULL OR student_id = sqlc.narg(student_id))
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: CreateStudentCardPortalLoginAttempt :one
INSERT INTO student_card_portal_login_attempts (card_id, student_id, challenge, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, NOW() + ($6::INT * INTERVAL '1 second'))
RETURNING *;

-- name: GetStartedStudentCardPortalLoginAttempt :one
SELECT * FROM student_card_portal_login_attempts
WHERE challenge = $1 AND status = 'started' AND expires_at > NOW();

-- name: UpdateStudentCardPortalLoginAttemptStatus :one
UPDATE student_card_portal_login_attempts
SET status = $2,
    completed_at = CASE WHEN $2 IN ('completed','failed','expired') THEN NOW() ELSE completed_at END
WHERE id = $1
RETURNING *;

-- name: CompleteStudentCardPortalLoginAttempt :one
UPDATE student_card_portal_login_attempts
SET status = 'completed', completed_at = NOW()
WHERE challenge = $1 AND status = 'started' AND expires_at > NOW()
RETURNING *;

-- name: IncrementStudentPortalPINFailure :one
UPDATE student_portal_pin_credentials
SET failed_attempts = failed_attempts + 1,
    locked_until = CASE WHEN failed_attempts + 1 >= 5 THEN NOW() + INTERVAL '15 minutes' ELSE locked_until END,
    updated_at = NOW()
WHERE student_id = $1
RETURNING *;

-- name: ResetStudentPortalPINFailure :one
UPDATE student_portal_pin_credentials
SET failed_attempts = 0,
    locked_until = NULL,
    updated_at = NOW()
WHERE student_id = $1
RETURNING *;

-- name: GetStudentPortalPINCredential :one
SELECT * FROM student_portal_pin_credentials
WHERE student_id = $1 AND is_enabled = TRUE AND (locked_until IS NULL OR locked_until <= NOW());

-- name: InsertStudentActivityAttendanceScan :one
INSERT INTO student_activity_attendance_scans (card_id, student_id, activity_code, scan_type, scanned_by_user_id, metadata)
VALUES ($1, $2, $3, $4, sqlc.narg(scanned_by_user_id), $5)
RETURNING *;

-- name: ValidateCardForCBT :many
SELECT p.id AS participant_id, p.session_id, p.student_id, es.title AS session_title, es.status AS session_status,
       es.scheduled_start, es.scheduled_end
FROM cbt_exam_participants p
JOIN cbt_exam_sessions es ON es.id = p.session_id
JOIN student_id_cards c ON c.student_id = p.student_id
WHERE c.token_hash = $1 AND c.status = 'active'
  AND es.status IN ('scheduled','active')
  AND es.scheduled_end >= NOW() - INTERVAL '2 hours'
ORDER BY es.scheduled_start ASC;
