-- name: ListCbtExamRooms :many
SELECT
  r.id, r.session_id, r.room_name, r.capacity, r.created_at,
  COUNT(p.id)::int AS participant_count
FROM cbt_exam_rooms r
LEFT JOIN cbt_exam_participants p ON p.room_id = r.id
WHERE r.session_id = $1
GROUP BY r.id
ORDER BY r.room_name ASC;

-- name: CreateCbtExamRoom :one
INSERT INTO cbt_exam_rooms (session_id, room_name, capacity)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteCbtExamRoom :exec
DELETE FROM cbt_exam_rooms WHERE id = $1;

-- name: GetCbtExamRoom :one
SELECT r.id, r.session_id, r.room_name, r.capacity, r.created_at
FROM cbt_exam_rooms r
WHERE r.id = $1;
