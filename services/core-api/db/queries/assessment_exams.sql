-- name: ListAssessmentExams :many
SELECT
  e.*,
  COUNT(DISTINCT s.id)::bigint AS session_count,
  COUNT(DISTINCT r.id)::bigint AS room_count,
  COUNT(DISTINCT p.id)::bigint AS participant_count
FROM assessment_exams e
LEFT JOIN assessment_sessions s ON s.exam_id = e.id
LEFT JOIN assessment_rooms r ON r.session_id = s.id
LEFT JOIN assessment_participants p ON p.session_id = s.id
WHERE (sqlc.arg(status_filter)::text = '' OR e.status = sqlc.arg(status_filter)::text)
  AND (sqlc.arg(search)::text = '' OR e.title ILIKE '%' || sqlc.arg(search)::text || '%')
GROUP BY e.id
ORDER BY e.created_at DESC
LIMIT LEAST(GREATEST(sqlc.arg(limit_count)::int, 1), 100)
OFFSET GREATEST(sqlc.arg(offset_count)::int, 0);

-- name: GetAssessmentExam :one
SELECT
  e.*,
  COUNT(DISTINCT s.id)::bigint AS session_count,
  COUNT(DISTINCT r.id)::bigint AS room_count,
  COUNT(DISTINCT p.id)::bigint AS participant_count,
  COUNT(DISTINCT c.id)::bigint AS card_count
FROM assessment_exams e
LEFT JOIN assessment_sessions s ON s.exam_id = e.id
LEFT JOIN assessment_rooms r ON r.session_id = s.id
LEFT JOIN assessment_participants p ON p.session_id = s.id
LEFT JOIN assessment_access_cards c ON c.session_id = s.id
WHERE e.id = sqlc.arg(id)
GROUP BY e.id;

-- name: CreateAssessmentExam :one
INSERT INTO assessment_exams (
  title,
  subject_id,
  grade_level,
  status,
  starts_at,
  ends_at,
  created_by
) VALUES (
  sqlc.arg(title),
  sqlc.narg(subject_id),
  sqlc.narg(grade_level),
  'draft',
  sqlc.narg(starts_at),
  sqlc.narg(ends_at),
  sqlc.narg(created_by)
)
RETURNING *;

-- name: UpdateAssessmentExam :one
UPDATE assessment_exams
SET title = sqlc.arg(title),
    subject_id = sqlc.narg(subject_id),
    grade_level = sqlc.narg(grade_level),
    status = sqlc.arg(status),
    starts_at = sqlc.narg(starts_at),
    ends_at = sqlc.narg(ends_at),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: GetFirstAssessmentSessionByExam :one
SELECT *
FROM assessment_sessions
WHERE exam_id = sqlc.arg(exam_id)
ORDER BY created_at ASC
LIMIT 1;

-- name: CreateAssessmentSession :one
INSERT INTO assessment_sessions (
  exam_id,
  title,
  starts_at,
  ends_at,
  status
) VALUES (
  sqlc.arg(exam_id),
  sqlc.arg(title),
  sqlc.narg(starts_at),
  sqlc.narg(ends_at),
  'draft'
)
RETURNING *;

-- name: CreateAssessmentRoom :one
INSERT INTO assessment_rooms (
  session_id,
  code,
  name,
  capacity,
  status
) VALUES (
  sqlc.arg(session_id),
  sqlc.arg(code),
  sqlc.arg(name),
  sqlc.arg(capacity),
  'draft'
)
ON CONFLICT (session_id, code) DO UPDATE
SET name = EXCLUDED.name,
    capacity = EXCLUDED.capacity,
    updated_at = now()
RETURNING *;

-- name: CountAssessmentRoomsByExam :one
SELECT COUNT(DISTINCT r.id)::bigint
FROM assessment_sessions s
JOIN assessment_rooms r ON r.session_id = s.id
WHERE s.exam_id = sqlc.arg(exam_id);

-- name: CountAssessmentParticipantsByExam :one
SELECT COUNT(DISTINCT p.id)::bigint
FROM assessment_sessions s
JOIN assessment_participants p ON p.session_id = s.id
WHERE s.exam_id = sqlc.arg(exam_id);

-- name: CountAssessmentCardsByExam :one
SELECT COUNT(DISTINCT c.id)::bigint
FROM assessment_sessions s
JOIN assessment_access_cards c ON c.session_id = s.id
WHERE s.exam_id = sqlc.arg(exam_id);
