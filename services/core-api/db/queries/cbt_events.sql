-- name: ListCbtExamEvents :many
SELECT
  e.id, e.title, e.exam_type, e.scope, e.status,
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
  e.id, e.title, e.exam_type, e.scope, e.status,
  e.academic_year_id,
  ay.name AS academic_year_name,
  e.created_at, e.updated_at
FROM cbt_exam_events e
LEFT JOIN academic_years ay ON ay.id = e.academic_year_id
WHERE e.id = $1;

-- name: CreateCbtExamEvent :one
INSERT INTO cbt_exam_events (title, exam_type, scope, academic_year_id, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateCbtExamEventStatus :one
UPDATE cbt_exam_events
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCbtExamEvent :one
UPDATE cbt_exam_events
SET title = $2, exam_type = $3, scope = $4, academic_year_id = $5, updated_at = NOW()
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
