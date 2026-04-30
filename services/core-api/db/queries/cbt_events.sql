-- name: ListCbtExamEvents :many
SELECT
  e.id, e.title, e.exam_type, e.scope, e.target_levels, e.status,
  e.academic_year_id,
  ay.name AS academic_year_name,
  e.created_at, e.updated_at,
  COUNT(s.id)::int AS session_count
FROM cbt_exam_events e
LEFT JOIN academic_years ay ON ay.id = e.academic_year_id
LEFT JOIN cbt_exam_sessions s ON s.event_id = e.id
GROUP BY e.id, ay.name
ORDER BY e.created_at DESC;

-- name: GetCbtExamEvent :one
SELECT
  e.id, e.title, e.exam_type, e.scope, e.target_levels, e.status,
  e.academic_year_id,
  ay.name AS academic_year_name,
  e.created_at, e.updated_at
FROM cbt_exam_events e
LEFT JOIN academic_years ay ON ay.id = e.academic_year_id
WHERE e.id = $1;

-- name: CreateCbtExamEvent :one
INSERT INTO cbt_exam_events (title, exam_type, scope, target_levels, academic_year_id, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateCbtExamEventStatus :one
UPDATE cbt_exam_events
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCbtExamEvent :one
UPDATE cbt_exam_events
SET title = $2, exam_type = $3, scope = $4, target_levels = $5, academic_year_id = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCbtExamEvent :exec
DELETE FROM cbt_exam_events WHERE id = $1 AND status = 'draft';

-- name: GetEventResults :many
SELECT 
    p.id AS participant_id,
    s.id AS session_id,
    s.title AS session_title,
    std.nis,
    std.nama AS student_nama,
    std.gender,
    c.code AS class_code,
    p.score,
    p.submitted_at
FROM cbt_exam_participants p
JOIN cbt_exam_sessions s ON s.id = p.session_id
JOIN students std ON std.id = p.student_id
LEFT JOIN school_classes c ON c.id = std.class_id
WHERE s.event_id = $1
ORDER BY std.nama ASC, s.scheduled_start ASC;

-- name: GetEventExamCards :many
SELECT
    e.id AS event_id,
    e.title AS event_title,
    e.exam_type,
    e.scope AS event_scope,
    e.target_levels,
    COALESCE(ay.name, '') AS academic_year_name,
    s.id AS session_id,
    s.title AS session_title,
    s.scheduled_start,
    p.id AS participant_id,
    p.token,
    p.seat_no,
    std.nis,
    std.nama AS student_nama,
    std.gender,
    COALESCE(c.code, '') AS class_code,
    COALESCE(r.room_name, '') AS room_name
FROM cbt_exam_events e
LEFT JOIN academic_years ay ON ay.id = e.academic_year_id
JOIN cbt_exam_sessions s ON s.event_id = e.id
JOIN cbt_exam_participants p ON p.session_id = s.id
JOIN students std ON std.id = p.student_id
LEFT JOIN school_classes c ON c.id = std.class_id
LEFT JOIN cbt_exam_rooms r ON r.id = p.room_id
WHERE e.id = $1
ORDER BY s.scheduled_start ASC, c.code ASC, std.nama ASC;
