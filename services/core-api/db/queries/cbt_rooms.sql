-- name: ListCbtExamRooms :many
WITH participant_counts AS (
  SELECT room_id, COUNT(*)::int AS participant_count
  FROM cbt_exam_participants
  WHERE session_id = $1 AND room_id IS NOT NULL
  GROUP BY room_id
),
proctor_counts AS (
  SELECT exam_room_id, COUNT(*)::int AS proctor_count
  FROM cbt_room_proctors
  GROUP BY exam_room_id
),
primary_proctors AS (
  SELECT DISTINCT ON (rp.exam_room_id)
    rp.exam_room_id, rp.employee_id AS primary_proctor_id, e.nama AS primary_proctor_name
  FROM cbt_room_proctors rp
  JOIN employees e ON e.id = rp.employee_id
  ORDER BY rp.exam_room_id, CASE rp.role WHEN 'utama' THEN 0 WHEN 'pendamping' THEN 1 ELSE 2 END, rp.assigned_at ASC
)
SELECT
  r.id, r.session_id, r.school_room_id,
  r.room_name, r.room_name_snapshot, r.capacity, r.capacity_override,
  r.room_token, r.status, r.is_locked, r.created_at, r.updated_at,
  r.allow_web_fallback, r.web_fallback_enabled_at, r.web_fallback_enabled_by,
  r.web_fallback_reason, r.web_fallback_disabled_at,
  COALESCE(sr.code, '') AS school_room_code,
  COALESCE(sr.name, '') AS school_room_name,
  COALESCE(sr.building, '') AS school_room_building,
  COALESCE(sr.location_note, '') AS school_room_location_note,
  COALESCE(sr.exam_capacity, 0)::int AS school_room_exam_capacity,
  COALESCE(sr.condition, '') AS school_room_condition,
  COALESCE(sr.is_exam_eligible, false) AS school_room_exam_eligible,
  COALESCE(pc.participant_count, 0)::int AS participant_count,
  COALESCE(prc.proctor_count, 0)::int AS proctor_count,
  pp.primary_proctor_id,
  COALESCE(pp.primary_proctor_name, '') AS primary_proctor_name
FROM cbt_exam_rooms r
LEFT JOIN school_rooms sr ON sr.id = r.school_room_id
LEFT JOIN participant_counts pc ON pc.room_id = r.id
LEFT JOIN proctor_counts prc ON prc.exam_room_id = r.id
LEFT JOIN primary_proctors pp ON pp.exam_room_id = r.id
WHERE r.session_id = $1
ORDER BY r.room_name ASC;

-- name: CreateCbtExamRoom :one
WITH room_seq AS (
  SELECT COALESCE(COUNT(*), 0)::int + 1 AS ordinal
  FROM cbt_exam_rooms
  WHERE session_id = sqlc.arg(session_id)
), generated AS (
  SELECT
    'R' || LPAD(room_seq.ordinal::text, 2, '0') || '-' || UPPER(SUBSTRING(encode(gen_random_bytes(3), 'hex') FROM 1 FOR 4)) AS room_token
  FROM room_seq
)
INSERT INTO cbt_exam_rooms (
  session_id, school_room_id, room_name, room_name_snapshot, capacity,
  room_token, room_token_hash, room_token_hash_version, room_token_generated_at
)
SELECT
  sqlc.arg(session_id), sqlc.arg(school_room_id), sqlc.arg(room_name),
  COALESCE(NULLIF(sqlc.arg(room_name_snapshot)::TEXT, ''), sqlc.arg(room_name)::TEXT),
  sqlc.arg(capacity),
  generated.room_token, encode(digest(generated.room_token, 'sha256'), 'hex'), 1, NOW()
FROM generated
RETURNING *;

-- name: UpdateCbtRoomWebFallbackPolicy :one
UPDATE cbt_exam_rooms
SET allow_web_fallback = sqlc.arg(allow_web_fallback),
    web_fallback_enabled_at = CASE
      WHEN sqlc.arg(allow_web_fallback)::boolean THEN COALESCE(web_fallback_enabled_at, NOW())
      ELSE web_fallback_enabled_at
    END,
    web_fallback_enabled_by = CASE
      WHEN sqlc.arg(allow_web_fallback)::boolean THEN sqlc.arg(actor_user_id)
      ELSE web_fallback_enabled_by
    END,
    web_fallback_reason = sqlc.arg(reason),
    web_fallback_disabled_at = CASE
      WHEN sqlc.arg(allow_web_fallback)::boolean THEN NULL
      ELSE NOW()
    END,
    updated_at = NOW()
WHERE session_id = sqlc.arg(session_id)
  AND id = sqlc.arg(room_id)
RETURNING *;

-- name: DeleteCbtExamRoom :exec
DELETE FROM cbt_exam_rooms WHERE id = $1;

-- name: GetCbtExamRoom :one
SELECT *
FROM cbt_exam_rooms
WHERE id = $1;

-- name: GetCbtRoomProctorDashboard :one
SELECT
  r.id,
  r.session_id,
  r.school_room_id,
  r.room_name,
  r.room_name_snapshot,
  r.capacity,
  r.capacity_override,
  r.room_token,
  r.status,
  r.is_locked,
  r.created_at,
  r.updated_at,
  r.allow_web_fallback,
  r.web_fallback_enabled_at,
  r.web_fallback_enabled_by,
  r.web_fallback_reason,
  r.web_fallback_disabled_at,
  s.title AS session_title,
  s.status AS session_status,
  s.scheduled_start,
  s.scheduled_end,
  p.title AS package_title,
  p.duration_minutes,
  COALESCE(sr.code, '') AS school_room_code,
  COALESCE(sr.name, '') AS school_room_name,
  COALESCE(sr.building, '') AS school_room_building,
  COALESCE(sr.location_note, '') AS school_room_location_note,
  COUNT(ep.id)::int AS participant_count,
  COUNT(ep.id) FILTER (WHERE ep.submitted_at IS NOT NULL)::int AS submitted_count,
  COUNT(ep.id) FILTER (WHERE ep.last_heartbeat IS NOT NULL AND ep.last_heartbeat > NOW() - INTERVAL '2 minutes')::int AS online_count,
  COUNT(ep.id) FILTER (WHERE ep.suspicious_flag = TRUE)::int AS suspicious_count,
  COUNT(ep.id) FILTER (WHERE ep.seat_no IS NULL)::int AS missing_seat_count
FROM cbt_exam_rooms r
JOIN cbt_exam_sessions s ON s.id = r.session_id
JOIN cbt_packages p ON p.id = s.package_id
LEFT JOIN school_rooms sr ON sr.id = r.school_room_id
LEFT JOIN cbt_exam_participants ep ON ep.room_id = r.id AND ep.session_id = r.session_id
WHERE r.id = $1
GROUP BY r.id, s.title, s.status, s.scheduled_start, s.scheduled_end, p.title, p.duration_minutes,
         sr.code, sr.name, sr.building, sr.location_note;

-- name: GetCbtRoomHandover :one
WITH room_stats AS (
  SELECT
    ep.room_id,
    COUNT(*)::int AS participant_count,
    COUNT(*) FILTER (WHERE ep.submitted_at IS NOT NULL)::int AS submitted_count,
    COUNT(*) FILTER (WHERE ep.suspicious_flag = TRUE)::int AS suspicious_count,
    COUNT(*) FILTER (WHERE ep.seat_no IS NULL)::int AS missing_seat_count
  FROM cbt_exam_participants ep
  WHERE ep.room_id = $1
  GROUP BY ep.room_id
)
SELECT
  r.id AS room_id,
  r.session_id,
  r.room_name,
  r.room_token,
  r.status AS room_status,
  r.is_locked AS room_is_locked,
  s.title AS session_title,
  s.status AS session_status,
  s.scheduled_start,
  s.scheduled_end,
  p.title AS package_title,
  COALESCE(rs.participant_count, 0)::int AS participant_count,
  COALESCE(rs.submitted_count, 0)::int AS submitted_count,
  COALESCE(rs.suspicious_count, 0)::int AS suspicious_count,
  COALESCE(rs.missing_seat_count, 0)::int AS missing_seat_count,
  h.id AS handover_id,
  COALESCE(h.attendance_checked, FALSE)::boolean AS attendance_checked,
  COALESCE(h.all_submitted_checked, FALSE)::boolean AS all_submitted_checked,
  COALESCE(h.device_issue_checked, FALSE)::boolean AS device_issue_checked,
  COALESCE(h.room_clean_checked, FALSE)::boolean AS room_clean_checked,
  COALESCE(h.token_returned_checked, FALSE)::boolean AS token_returned_checked,
  COALESCE(h.assets_returned_checked, FALSE)::boolean AS assets_returned_checked,
  COALESCE(h.incident_notes, '')::text AS incident_notes,
  COALESCE(h.operator_notes, '')::text AS operator_notes,
  COALESCE(h.handover_notes, '')::text AS handover_notes,
  h.locked_at,
  h.locked_by,
  h.updated_by,
  h.created_at AS handover_created_at,
  h.updated_at AS handover_updated_at
FROM cbt_exam_rooms r
JOIN cbt_exam_sessions s ON s.id = r.session_id
JOIN cbt_packages p ON p.id = s.package_id
LEFT JOIN room_stats rs ON rs.room_id = r.id
LEFT JOIN cbt_room_handovers h ON h.exam_room_id = r.id
WHERE r.id = $1;

-- name: GetCbtSessionOperationalRecap :one
WITH participant_stats AS (
  SELECT
    ep.session_id,
    COUNT(*)::int AS participant_count,
    COUNT(*) FILTER (WHERE ep.room_id IS NULL)::int AS unassigned_participant_count,
    COUNT(*) FILTER (WHERE ep.joined_at IS NOT NULL)::int AS joined_count,
    COUNT(*) FILTER (WHERE ep.submitted_at IS NOT NULL)::int AS submitted_count,
    COUNT(*) FILTER (WHERE ep.joined_at IS NULL AND ep.submitted_at IS NULL)::int AS no_show_count,
    COUNT(*) FILTER (WHERE ep.suspicious_flag = TRUE)::int AS suspicious_count,
    COALESCE(SUM(ep.app_switch_count), 0)::int AS app_switch_count,
    COALESCE(SUM(ep.screenshot_attempt), 0)::int AS screenshot_attempt_count
  FROM cbt_exam_participants ep
  WHERE ep.session_id = $1
  GROUP BY ep.session_id
),
room_stats AS (
  SELECT
    r.session_id,
    COUNT(*)::int AS room_count,
    COUNT(*) FILTER (WHERE h.locked_at IS NOT NULL)::int AS handover_locked_count,
    COUNT(*) FILTER (WHERE h.id IS NOT NULL AND h.locked_at IS NULL)::int AS handover_draft_count,
    COUNT(*) FILTER (WHERE h.id IS NULL)::int AS handover_missing_count,
    COUNT(*) FILTER (
      WHERE NULLIF(TRIM(COALESCE(h.incident_notes, '')), '') IS NOT NULL
         OR NULLIF(TRIM(COALESCE(h.operator_notes, '')), '') IS NOT NULL
    )::int AS incident_room_count
  FROM cbt_exam_rooms r
  LEFT JOIN cbt_room_handovers h ON h.exam_room_id = r.id
  WHERE r.session_id = $1
  GROUP BY r.session_id
),
event_stats AS (
  SELECT
    ep.session_id,
    COUNT(ev.id) FILTER (
      WHERE ev.event_type IN ('app_switch', 'screenshot_attempt', 'proctor_force_submit', 'proctor_reset_access', 'warning')
    )::int AS incident_event_count,
    COUNT(ev.id) FILTER (WHERE ev.event_type = 'proctor_force_submit')::int AS force_submit_count,
    COUNT(ev.id) FILTER (WHERE ev.event_type = 'proctor_reset_access')::int AS reset_access_count,
    COUNT(ev.id) FILTER (WHERE ev.event_type = 'app_switch')::int AS app_switch_event_count,
    COUNT(ev.id) FILTER (WHERE ev.event_type = 'screenshot_attempt')::int AS screenshot_event_count
  FROM cbt_exam_participants ep
  LEFT JOIN cbt_participant_events ev ON ev.participant_id = ep.id
  WHERE ep.session_id = $1
  GROUP BY ep.session_id
)
SELECT
  s.id AS session_id,
  s.title AS session_title,
  s.status AS session_status,
  s.scheduled_start,
  s.scheduled_end,
  COALESCE(rs.room_count, 0)::int AS room_count,
  COALESCE(rs.handover_locked_count, 0)::int AS handover_locked_count,
  COALESCE(rs.handover_draft_count, 0)::int AS handover_draft_count,
  COALESCE(rs.handover_missing_count, 0)::int AS handover_missing_count,
  COALESCE(rs.incident_room_count, 0)::int AS incident_room_count,
  COALESCE(ps.participant_count, 0)::int AS participant_count,
  COALESCE(ps.unassigned_participant_count, 0)::int AS unassigned_participant_count,
  COALESCE(ps.joined_count, 0)::int AS joined_count,
  COALESCE(ps.submitted_count, 0)::int AS submitted_count,
  COALESCE(ps.no_show_count, 0)::int AS no_show_count,
  COALESCE(ps.suspicious_count, 0)::int AS suspicious_count,
  COALESCE(ps.app_switch_count, 0)::int AS app_switch_count,
  COALESCE(ps.screenshot_attempt_count, 0)::int AS screenshot_attempt_count,
  COALESCE(es.incident_event_count, 0)::int AS incident_event_count,
  COALESCE(es.force_submit_count, 0)::int AS force_submit_count,
  COALESCE(es.reset_access_count, 0)::int AS reset_access_count,
  COALESCE(es.app_switch_event_count, 0)::int AS app_switch_event_count,
  COALESCE(es.screenshot_event_count, 0)::int AS screenshot_event_count
FROM cbt_exam_sessions s
LEFT JOIN participant_stats ps ON ps.session_id = s.id
LEFT JOIN room_stats rs ON rs.session_id = s.id
LEFT JOIN event_stats es ON es.session_id = s.id
WHERE s.id = $1;

-- name: ListCbtSessionRoomOperationalRecap :many
WITH participant_stats AS (
  SELECT
    ep.room_id,
    COUNT(*)::int AS participant_count,
    COUNT(*) FILTER (WHERE ep.joined_at IS NOT NULL)::int AS joined_count,
    COUNT(*) FILTER (WHERE ep.submitted_at IS NOT NULL)::int AS submitted_count,
    COUNT(*) FILTER (WHERE ep.joined_at IS NULL AND ep.submitted_at IS NULL)::int AS no_show_count,
    COUNT(*) FILTER (WHERE ep.suspicious_flag = TRUE)::int AS suspicious_count,
    COUNT(*) FILTER (WHERE ep.seat_no IS NULL)::int AS missing_seat_count,
    COALESCE(SUM(ep.app_switch_count), 0)::int AS app_switch_count,
    COALESCE(SUM(ep.screenshot_attempt), 0)::int AS screenshot_attempt_count
  FROM cbt_exam_participants ep
  WHERE ep.session_id = $1 AND ep.room_id IS NOT NULL
  GROUP BY ep.room_id
),
event_stats AS (
  SELECT
    ep.room_id,
    COUNT(ev.id) FILTER (
      WHERE ev.event_type IN ('app_switch', 'screenshot_attempt', 'proctor_force_submit', 'proctor_reset_access', 'warning')
    )::int AS incident_event_count,
    COUNT(ev.id) FILTER (WHERE ev.event_type = 'proctor_force_submit')::int AS force_submit_count,
    COUNT(ev.id) FILTER (WHERE ev.event_type = 'proctor_reset_access')::int AS reset_access_count
  FROM cbt_exam_participants ep
  LEFT JOIN cbt_participant_events ev ON ev.participant_id = ep.id
  WHERE ep.session_id = $1 AND ep.room_id IS NOT NULL
  GROUP BY ep.room_id
)
SELECT
  r.id AS room_id,
  r.session_id,
  r.room_name,
  r.room_token,
  r.status AS room_status,
  r.is_locked AS room_is_locked,
  COALESCE(ps.participant_count, 0)::int AS participant_count,
  COALESCE(ps.joined_count, 0)::int AS joined_count,
  COALESCE(ps.submitted_count, 0)::int AS submitted_count,
  COALESCE(ps.no_show_count, 0)::int AS no_show_count,
  COALESCE(ps.suspicious_count, 0)::int AS suspicious_count,
  COALESCE(ps.missing_seat_count, 0)::int AS missing_seat_count,
  COALESCE(ps.app_switch_count, 0)::int AS app_switch_count,
  COALESCE(ps.screenshot_attempt_count, 0)::int AS screenshot_attempt_count,
  COALESCE(es.incident_event_count, 0)::int AS incident_event_count,
  COALESCE(es.force_submit_count, 0)::int AS force_submit_count,
  COALESCE(es.reset_access_count, 0)::int AS reset_access_count,
  h.id AS handover_id,
  COALESCE(h.attendance_checked, FALSE)::boolean AS attendance_checked,
  COALESCE(h.all_submitted_checked, FALSE)::boolean AS all_submitted_checked,
  COALESCE(h.device_issue_checked, FALSE)::boolean AS device_issue_checked,
  COALESCE(h.room_clean_checked, FALSE)::boolean AS room_clean_checked,
  COALESCE(h.token_returned_checked, FALSE)::boolean AS token_returned_checked,
  COALESCE(h.assets_returned_checked, FALSE)::boolean AS assets_returned_checked,
  COALESCE(h.incident_notes, '')::text AS incident_notes,
  COALESCE(h.operator_notes, '')::text AS operator_notes,
  COALESCE(h.handover_notes, '')::text AS handover_notes,
  h.locked_at,
  h.updated_at AS handover_updated_at
FROM cbt_exam_rooms r
LEFT JOIN participant_stats ps ON ps.room_id = r.id
LEFT JOIN event_stats es ON es.room_id = r.id
LEFT JOIN cbt_room_handovers h ON h.exam_room_id = r.id
WHERE r.session_id = $1
ORDER BY
  CASE
    WHEN h.id IS NULL THEN 0
    WHEN h.locked_at IS NULL THEN 1
    ELSE 2
  END,
  COALESCE(es.incident_event_count, 0) DESC,
  r.room_name ASC;

-- name: UpsertCbtRoomHandover :one
INSERT INTO cbt_room_handovers (
  exam_room_id,
  attendance_checked,
  all_submitted_checked,
  device_issue_checked,
  room_clean_checked,
  token_returned_checked,
  assets_returned_checked,
  incident_notes,
  operator_notes,
  handover_notes,
  updated_by
)
VALUES (
  sqlc.arg(exam_room_id),
  sqlc.arg(attendance_checked),
  sqlc.arg(all_submitted_checked),
  sqlc.arg(device_issue_checked),
  sqlc.arg(room_clean_checked),
  sqlc.arg(token_returned_checked),
  sqlc.arg(assets_returned_checked),
  sqlc.arg(incident_notes),
  sqlc.arg(operator_notes),
  sqlc.arg(handover_notes),
  sqlc.arg(updated_by)
)
ON CONFLICT (exam_room_id) DO UPDATE
SET
  attendance_checked = EXCLUDED.attendance_checked,
  all_submitted_checked = EXCLUDED.all_submitted_checked,
  device_issue_checked = EXCLUDED.device_issue_checked,
  room_clean_checked = EXCLUDED.room_clean_checked,
  token_returned_checked = EXCLUDED.token_returned_checked,
  assets_returned_checked = EXCLUDED.assets_returned_checked,
  incident_notes = EXCLUDED.incident_notes,
  operator_notes = EXCLUDED.operator_notes,
  handover_notes = EXCLUDED.handover_notes,
  updated_by = EXCLUDED.updated_by,
  updated_at = NOW()
WHERE cbt_room_handovers.locked_at IS NULL
RETURNING *;

-- name: LockCbtRoomHandover :one
INSERT INTO cbt_room_handovers (
  exam_room_id,
  locked_at,
  locked_by,
  updated_by
)
VALUES (
  sqlc.arg(exam_room_id),
  NOW(),
  sqlc.arg(locked_by),
  sqlc.arg(locked_by)
)
ON CONFLICT (exam_room_id) DO UPDATE
SET
  locked_at = NOW(),
  locked_by = EXCLUDED.locked_by,
  updated_by = EXCLUDED.updated_by,
  updated_at = NOW()
WHERE cbt_room_handovers.locked_at IS NULL
RETURNING *;

-- name: ListCbtProctorRooms :many
WITH participant_stats AS (
  SELECT
    ep.room_id,
    COUNT(*)::int AS participant_count,
    COUNT(*) FILTER (WHERE ep.submitted_at IS NOT NULL)::int AS submitted_count,
    COUNT(*) FILTER (WHERE ep.last_heartbeat IS NOT NULL AND ep.last_heartbeat > NOW() - INTERVAL '2 minutes')::int AS online_count,
    COUNT(*) FILTER (WHERE ep.suspicious_flag = TRUE)::int AS suspicious_count,
    COUNT(*) FILTER (WHERE ep.seat_no IS NULL)::int AS missing_seat_count
  FROM cbt_exam_participants ep
  WHERE ep.room_id IS NOT NULL
  GROUP BY ep.room_id
),
proctor_rollup AS (
  SELECT
    rp.exam_room_id,
    COUNT(*)::int AS proctor_count,
    STRING_AGG(e.nama, ', ' ORDER BY CASE rp.role WHEN 'utama' THEN 0 WHEN 'pendamping' THEN 1 ELSE 2 END, e.nama ASC)::text AS proctor_names,
    COALESCE(
      MAX(CASE WHEN rp.employee_id = sqlc.arg(employee_id) THEN rp.role ELSE '' END),
      ''
    )::text AS actor_role
  FROM cbt_room_proctors rp
  JOIN employees e ON e.id = rp.employee_id
  GROUP BY rp.exam_room_id
)
SELECT
  r.id,
  r.session_id,
  r.school_room_id,
  r.room_name,
  r.room_name_snapshot,
  r.capacity,
  r.capacity_override,
  r.room_token,
  r.status,
  r.is_locked,
  r.created_at,
  r.updated_at,
  s.title AS session_title,
  s.status AS session_status,
  s.scheduled_start,
  s.scheduled_end,
  p.title AS package_title,
  p.duration_minutes,
  COALESCE(sr.code, '') AS school_room_code,
  COALESCE(sr.name, '') AS school_room_name,
  COALESCE(sr.building, '') AS school_room_building,
  COALESCE(sr.location_note, '') AS school_room_location_note,
  COALESCE(ps.participant_count, 0)::int AS participant_count,
  COALESCE(ps.submitted_count, 0)::int AS submitted_count,
  COALESCE(ps.online_count, 0)::int AS online_count,
  COALESCE(ps.suspicious_count, 0)::int AS suspicious_count,
  COALESCE(ps.missing_seat_count, 0)::int AS missing_seat_count,
  COALESCE(pr.proctor_count, 0)::int AS proctor_count,
  COALESCE(pr.proctor_names, '')::text AS proctor_names,
  COALESCE(pr.actor_role, '')::text AS actor_role
FROM cbt_exam_rooms r
JOIN cbt_exam_sessions s ON s.id = r.session_id
JOIN cbt_packages p ON p.id = s.package_id
LEFT JOIN school_rooms sr ON sr.id = r.school_room_id
LEFT JOIN participant_stats ps ON ps.room_id = r.id
LEFT JOIN proctor_rollup pr ON pr.exam_room_id = r.id
WHERE (
    sqlc.arg(include_all)::boolean
    OR EXISTS (
      SELECT 1
      FROM cbt_room_proctors actor_rp
      WHERE actor_rp.exam_room_id = r.id
        AND actor_rp.employee_id = sqlc.arg(employee_id)
    )
  )
  AND (s.status IN ('draft', 'scheduled', 'active') OR s.scheduled_end >= NOW() - INTERVAL '7 days')
ORDER BY CASE s.status WHEN 'active' THEN 0 WHEN 'scheduled' THEN 1 WHEN 'draft' THEN 2 ELSE 3 END,
         s.scheduled_start ASC,
         r.room_name ASC;

-- name: ListCbtRoomProctors :many
SELECT rp.id, rp.exam_room_id, rp.employee_id, COALESCE(e.nip, '')::text AS nip, e.nama,
       rp.role, rp.assigned_by, rp.assigned_at
FROM cbt_room_proctors rp
JOIN employees e ON e.id = rp.employee_id
WHERE rp.exam_room_id = $1
ORDER BY CASE rp.role WHEN 'utama' THEN 0 WHEN 'pendamping' THEN 1 ELSE 2 END, e.nama ASC;

-- name: DeleteCbtRoomProctorsByRoom :exec
DELETE FROM cbt_room_proctors WHERE exam_room_id = $1;

-- name: CreateCbtRoomProctor :one
INSERT INTO cbt_room_proctors (exam_room_id, employee_id, role, assigned_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: HasSessionRoomProctor :one
SELECT EXISTS (
  SELECT 1
  FROM cbt_room_proctors rp
  JOIN cbt_exam_rooms r ON r.id = rp.exam_room_id
  WHERE r.session_id = sqlc.arg(session_id)
    AND r.id = sqlc.arg(room_id)
    AND rp.employee_id = sqlc.arg(employee_id)
)::boolean;

-- name: HasSessionRoomParticipant :one
SELECT EXISTS (
  SELECT 1
  FROM cbt_exam_participants ep
  WHERE ep.session_id = sqlc.arg(session_id)
    AND ep.room_id = sqlc.arg(room_id)
    AND ep.id = sqlc.arg(participant_id)
)::boolean;

-- name: GetCbtSessionRoomReadiness :one
WITH room_stats AS (
  SELECT
    COUNT(*)::int AS room_count,
    COALESCE(SUM(COALESCE(capacity_override, capacity)), 0)::int AS total_capacity
  FROM cbt_exam_rooms er
  WHERE er.session_id = sqlc.arg(target_session_id)
),
participant_stats AS (
  SELECT
    COUNT(*)::int AS participant_count,
    COUNT(*) FILTER (WHERE room_id IS NOT NULL)::int AS assigned_participant_count,
    COUNT(*) FILTER (WHERE room_id IS NULL)::int AS unassigned_participant_count,
    COUNT(*) FILTER (WHERE room_id IS NOT NULL AND seat_no IS NULL)::int AS missing_seat_count
  FROM cbt_exam_participants ep
  WHERE ep.session_id = sqlc.arg(target_session_id)
),
room_proctor_counts AS (
  SELECT r.id, COUNT(rp.id)::int AS proctor_count
  FROM cbt_exam_rooms r
  LEFT JOIN cbt_room_proctors rp ON rp.exam_room_id = r.id
  WHERE r.session_id = sqlc.arg(target_session_id)
  GROUP BY r.id
),
proctor_stats AS (
  SELECT
    COUNT(*) FILTER (WHERE proctor_count = 0)::int AS rooms_without_proctor,
    COALESCE(SUM(proctor_count), 0)::int AS proctor_assignment_count
  FROM room_proctor_counts
),
room_capacity_stats AS (
  SELECT
    r.id,
    r.room_name,
    COALESCE(r.capacity_override, r.capacity)::int AS effective_capacity,
    COUNT(ep.id)::int AS participant_count
  FROM cbt_exam_rooms r
  LEFT JOIN cbt_exam_participants ep ON ep.room_id = r.id AND ep.session_id = r.session_id
  WHERE r.session_id = sqlc.arg(target_session_id)
  GROUP BY r.id, r.room_name, r.capacity_override, r.capacity
),
over_capacity_stats AS (
  SELECT
    COUNT(*) FILTER (WHERE participant_count > effective_capacity)::int AS over_capacity_room_count,
    COALESCE(
      jsonb_agg(
        jsonb_build_object(
          'room_id', id,
          'room_name', room_name,
          'participant_count', participant_count,
          'effective_capacity', effective_capacity
        ) ORDER BY room_name
      ) FILTER (WHERE participant_count > effective_capacity),
      '[]'::jsonb
    ) AS over_capacity_rooms
  FROM room_capacity_stats
),
physical_room_stats AS (
  SELECT
    COUNT(*) FILTER (WHERE r.school_room_id IS NOT NULL AND COALESCE(sr.network_ready, FALSE) = FALSE)::int AS network_not_ready_room_count,
    COUNT(*) FILTER (WHERE r.school_room_id IS NOT NULL AND COALESCE(sr.power_ready, FALSE) = FALSE)::int AS power_not_ready_room_count
  FROM cbt_exam_rooms r
  LEFT JOIN school_rooms sr ON sr.id = r.school_room_id
  WHERE r.session_id = sqlc.arg(target_session_id)
)
SELECT
  rs.room_count,
  rs.total_capacity,
  ps.participant_count,
  ps.assigned_participant_count,
  ps.unassigned_participant_count,
  ps.missing_seat_count,
  COALESCE(prs.rooms_without_proctor, 0)::int AS rooms_without_proctor,
  COALESCE(prs.proctor_assignment_count, 0)::int AS proctor_assignment_count,
  COALESCE(ocs.over_capacity_room_count, 0)::int AS over_capacity_room_count,
  COALESCE(ocs.over_capacity_rooms, '[]'::jsonb) AS over_capacity_rooms,
  COALESCE(phys.network_not_ready_room_count, 0)::int AS network_not_ready_room_count,
  COALESCE(phys.power_not_ready_room_count, 0)::int AS power_not_ready_room_count
FROM room_stats rs, participant_stats ps, proctor_stats prs, over_capacity_stats ocs, physical_room_stats phys;

-- name: GetCbtExamRoomSetupContext :one
SELECT
  r.id,
  r.session_id,
  r.school_room_id,
  r.room_name,
  r.room_name_snapshot,
  r.capacity,
  r.capacity_override,
  r.room_token,
  r.status,
  r.is_locked,
  r.created_at,
  r.updated_at,
  s.status AS session_status,
  s.scheduled_start,
  s.scheduled_end
FROM cbt_exam_rooms r
JOIN cbt_exam_sessions s ON s.id = r.session_id
WHERE r.id = $1;

-- name: GetCbtExamRoomSetupUsage :one
SELECT
  COUNT(DISTINCT ep.id)::int AS participant_count,
  COUNT(DISTINCT rp.id)::int AS proctor_count,
  COUNT(DISTINCT h.id)::int AS handover_count
FROM cbt_exam_rooms r
LEFT JOIN cbt_exam_participants ep ON ep.room_id = r.id AND ep.session_id = r.session_id
LEFT JOIN cbt_room_proctors rp ON rp.exam_room_id = r.id
LEFT JOIN cbt_room_handovers h ON h.exam_room_id = r.id
WHERE r.id = $1
GROUP BY r.id;

-- name: ClearParticipantSeatsForSession :exec
UPDATE cbt_exam_participants
SET seat_no = NULL
WHERE session_id = $1
  AND room_id IS NOT NULL;

-- name: HasOverlappingCbtRoomProctor :one
SELECT EXISTS (
  SELECT 1
  FROM cbt_room_proctors rp
  JOIN cbt_exam_rooms r ON r.id = rp.exam_room_id
  JOIN cbt_exam_sessions s ON s.id = r.session_id
  JOIN cbt_exam_sessions target_session ON target_session.id = sqlc.arg(session_id)
  WHERE rp.employee_id = sqlc.arg(employee_id)
    AND r.id <> sqlc.arg(exam_room_id)
    AND s.status IN ('draft', 'scheduled', 'active')
    AND s.scheduled_start < target_session.scheduled_end
    AND target_session.scheduled_start < s.scheduled_end
)::boolean;
