-- name: GetCbtParticipantRiskForUpdate :one
SELECT
  ep.id,
  ep.session_id,
  ep.room_id,
  ep.violation_count,
  ep.risk_score,
  ep.risk_level,
  ep.locked_at,
  ep.locked_reason,
  ep.app_switch_count,
  ep.screenshot_attempt,
  ep.last_local_save_at,
  ep.last_synced_at,
  ep.pending_answer_count,
  ep.sync_state
FROM cbt_exam_participants ep
WHERE ep.id = $1
FOR UPDATE;

-- name: GetCbtParticipantProctorScope :one
SELECT
  ep.id AS participant_id,
  ep.session_id,
  ep.room_id,
  ep.student_id,
  s.nis,
  s.nama,
  COALESCE(r.room_name, '') AS room_name
FROM cbt_exam_participants ep
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.session_id = sqlc.arg(session_id)
  AND ep.id = sqlc.arg(participant_id);

-- name: UpdateCbtParticipantProctorRisk :one
UPDATE cbt_exam_participants
SET violation_count = sqlc.arg(violation_count),
    risk_score = sqlc.arg(risk_score),
    risk_level = sqlc.arg(risk_level),
    locked_at = sqlc.arg(locked_at),
    locked_reason = sqlc.arg(locked_reason),
    suspicious_flag = CASE WHEN sqlc.arg(suspicious_flag)::boolean THEN TRUE ELSE suspicious_flag END,
    app_switch_count = app_switch_count + sqlc.arg(app_switch_increment)::integer,
    screenshot_attempt = screenshot_attempt + sqlc.arg(screenshot_increment)::integer,
    last_local_save_at = sqlc.arg(last_local_save_at),
    last_synced_at = sqlc.arg(last_synced_at),
    pending_answer_count = sqlc.arg(pending_answer_count),
    sync_state = sqlc.arg(sync_state)
WHERE id = sqlc.arg(id)
RETURNING id, violation_count, risk_score, risk_level, locked_at, locked_reason,
  last_local_save_at, last_synced_at, pending_answer_count, sync_state;

-- name: CreateCbtParticipantProctorEvent :one
INSERT INTO cbt_participant_events (
  participant_id,
  event_type,
  event_data,
  severity,
  category,
  risk_delta,
  dedup_key,
  correlation_id,
  original_event_at,
  requires_note,
  actor_user_id,
  actor_username_snapshot,
  actor_employee_id,
  request_id,
  source_ip
)
VALUES (
  sqlc.arg(participant_id),
  sqlc.arg(event_type),
  sqlc.arg(event_data),
  sqlc.arg(severity),
  sqlc.arg(category),
  sqlc.arg(risk_delta),
  sqlc.arg(dedup_key),
  sqlc.arg(correlation_id),
  sqlc.arg(original_event_at),
  sqlc.arg(requires_note),
  sqlc.narg(actor_user_id),
  sqlc.arg(actor_username_snapshot),
  sqlc.narg(actor_employee_id),
  sqlc.arg(request_id),
  sqlc.arg(source_ip)
)
RETURNING id, participant_id, event_type, event_data, created_at, severity, category, risk_delta, dedup_key,
  acknowledged_at, acknowledged_by, acknowledge_note, requires_note;

-- name: GetRecentCbtProctorEventByDedupKey :one
SELECT id, participant_id, event_type, event_data, created_at, severity, category, risk_delta, dedup_key,
  acknowledged_at, acknowledged_by, acknowledge_note, requires_note
FROM cbt_participant_events
WHERE participant_id = sqlc.arg(participant_id)
  AND dedup_key = sqlc.arg(dedup_key)
  AND dedup_key <> ''
  AND created_at >= sqlc.arg(since_at)
ORDER BY created_at DESC
LIMIT 1;

-- name: GetCbtProctorEventScope :one
SELECT
  ev.id,
  ev.participant_id,
  ep.session_id,
  ep.room_id,
  ev.event_type,
  ev.severity,
  ev.category,
  ev.risk_delta,
  ev.acknowledged_at,
  ev.acknowledged_by,
  ev.acknowledge_note,
  ev.requires_note,
  ev.created_at,
  s.nis,
  s.nama,
  COALESCE(r.room_name, '') AS room_name
FROM cbt_participant_events ev
JOIN cbt_exam_participants ep ON ep.id = ev.participant_id
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ev.id = $1;

-- name: AcknowledgeCbtProctorEvent :one
UPDATE cbt_participant_events
SET acknowledged_at = COALESCE(acknowledged_at, NOW()),
    acknowledged_by = COALESCE(acknowledged_by, sqlc.arg(acknowledged_by)),
    acknowledge_note = CASE
      WHEN acknowledge_note = '' THEN sqlc.arg(acknowledge_note)
      ELSE acknowledge_note
    END
WHERE id = sqlc.arg(id)
RETURNING id, participant_id, event_type, event_data, created_at, severity, category, risk_delta, dedup_key,
  acknowledged_at, acknowledged_by, acknowledge_note, requires_note;

-- name: ListCbtProctorEventsBySession :many
SELECT
  ev.id,
  ev.participant_id,
  ep.student_id,
  s.nis,
  s.nama,
  ep.session_id,
  ep.room_id,
  COALESCE(r.room_name, '') AS room_name,
  ev.event_type,
  ev.severity,
  ev.category,
  ev.risk_delta,
  ev.dedup_key,
  ev.event_data,
  ev.created_at,
  ev.acknowledged_at,
  ev.acknowledged_by,
  ev.acknowledge_note,
  ev.requires_note,
  ev.actor_username_snapshot
FROM cbt_participant_events ev
JOIN cbt_exam_participants ep ON ep.id = ev.participant_id
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.session_id = sqlc.arg(session_id)
  AND (sqlc.narg(participant_id)::uuid IS NULL OR ev.participant_id = sqlc.narg(participant_id)::uuid)
ORDER BY ev.created_at DESC
LIMIT sqlc.arg(limit_count);

-- name: ListCbtProctorEventsByRoom :many
SELECT
  ev.id,
  ev.participant_id,
  ep.student_id,
  s.nis,
  s.nama,
  ep.session_id,
  ep.room_id,
  COALESCE(r.room_name, '') AS room_name,
  ev.event_type,
  ev.severity,
  ev.category,
  ev.risk_delta,
  ev.dedup_key,
  ev.event_data,
  ev.created_at,
  ev.acknowledged_at,
  ev.acknowledged_by,
  ev.acknowledge_note,
  ev.requires_note,
  ev.actor_username_snapshot
FROM cbt_participant_events ev
JOIN cbt_exam_participants ep ON ep.id = ev.participant_id
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.session_id = sqlc.arg(session_id)
  AND ep.room_id = sqlc.arg(room_id)
  AND (sqlc.narg(participant_id)::uuid IS NULL OR ev.participant_id = sqlc.narg(participant_id)::uuid)
ORDER BY ev.created_at DESC
LIMIT sqlc.arg(limit_count);

-- name: CreateCbtProctorAction :one
INSERT INTO cbt_proctor_actions (
  session_id,
  room_id,
  participant_id,
  event_id,
  action_type,
  reason,
  notes,
  actor_user_id,
  actor_username_snapshot,
  actor_employee_id,
  request_id,
  source_ip
)
VALUES (
  sqlc.arg(session_id),
  sqlc.narg(room_id),
  sqlc.narg(participant_id),
  sqlc.narg(event_id),
  sqlc.arg(action_type),
  sqlc.arg(reason),
  sqlc.arg(notes),
  sqlc.arg(actor_user_id),
  sqlc.arg(actor_username_snapshot),
  sqlc.narg(actor_employee_id),
  sqlc.arg(request_id),
  sqlc.arg(source_ip)
)
RETURNING *;

-- name: ListCbtProctorActionsBySession :many
SELECT *
FROM cbt_proctor_actions
WHERE session_id = sqlc.arg(session_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count);

-- name: ListCbtProctorActionsByRoom :many
SELECT *
FROM cbt_proctor_actions
WHERE session_id = sqlc.arg(session_id)
  AND room_id = sqlc.arg(room_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count);

-- name: UnlockParticipantAccessForProctor :one
UPDATE cbt_exam_participants
SET locked_at = NULL,
    locked_reason = NULL,
    risk_level = CASE
      WHEN risk_score >= 50 OR violation_count >= 2 THEN 'high'
      WHEN risk_score >= 20 OR violation_count >= 1 THEN 'warning'
      ELSE 'normal'
    END,
    suspicious_flag = TRUE
WHERE id = $1
RETURNING id, violation_count, risk_score, risk_level, locked_at, locked_reason,
  last_local_save_at, last_synced_at, pending_answer_count, sync_state;

-- name: HoldParticipantAccessForProctor :one
UPDATE cbt_exam_participants
SET locked_at = COALESCE(locked_at, NOW()),
    locked_reason = sqlc.arg(locked_reason),
    risk_level = 'locked',
    suspicious_flag = TRUE
WHERE id = sqlc.arg(id)
RETURNING id, violation_count, risk_score, risk_level, locked_at, locked_reason,
  last_local_save_at, last_synced_at, pending_answer_count, sync_state;

-- name: ResetParticipantDeviceBindingForProctor :one
UPDATE cbt_exam_participants
SET device_fingerprint = NULL,
    login_ip = NULL,
    last_heartbeat = NULL,
    locked_at = NULL,
    locked_reason = NULL,
    risk_level = CASE
      WHEN risk_score >= 50 OR violation_count >= 2 THEN 'high'
      WHEN risk_score >= 20 OR violation_count >= 1 THEN 'warning'
      ELSE 'normal'
    END,
    suspicious_flag = TRUE
WHERE id = $1
RETURNING id, violation_count, risk_score, risk_level, locked_at, locked_reason,
  last_local_save_at, last_synced_at, pending_answer_count, sync_state;
