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
INSERT INTO cbt_exam_rooms (
  session_id, school_room_id, room_name, room_name_snapshot, capacity
)
VALUES (
  sqlc.arg(session_id), sqlc.arg(school_room_id), sqlc.arg(room_name),
  COALESCE(NULLIF(sqlc.arg(room_name_snapshot)::TEXT, ''), sqlc.arg(room_name)::TEXT),
  sqlc.arg(capacity)
)
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
SELECT rp.id, rp.exam_room_id, rp.employee_id, e.nip, e.nama,
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
    COALESCE(SUM(capacity), 0)::int AS total_capacity
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
)
SELECT
  rs.room_count,
  rs.total_capacity,
  ps.participant_count,
  ps.assigned_participant_count,
  ps.unassigned_participant_count,
  ps.missing_seat_count,
  COALESCE(prs.rooms_without_proctor, 0)::int AS rooms_without_proctor,
  COALESCE(prs.proctor_assignment_count, 0)::int AS proctor_assignment_count
FROM room_stats rs, participant_stats ps, proctor_stats prs;
