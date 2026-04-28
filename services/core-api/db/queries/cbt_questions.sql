-- name: ListCbtQuestions :many
SELECT q.id, q.subject_id, s.name AS subject_name, s.code AS subject_code,
       q.code, q.question_text, q.question_type, q.options,
       q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
       q.answer_key, q.explanation, q.difficulty, q.status, q.created_at, q.updated_at
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
ORDER BY q.created_at DESC;

-- name: GetCbtQuestion :one
SELECT id, subject_id, code, question_text, question_type, options,
       option_a, option_b, option_c, option_d, option_e,
       answer_key, explanation, difficulty, status, created_at, updated_at
FROM cbt_questions
WHERE id = $1;

-- name: CreateCbtQuestion :one
INSERT INTO cbt_questions (
  id, subject_id, code, question_text, question_type, options,
  option_a, option_b, option_c, option_d, option_e,
  answer_key, explanation, difficulty, status
)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: DeleteCbtQuestion :exec
DELETE FROM cbt_questions WHERE id = $1;

-- name: ListUngradedEssays :many
SELECT
  sa.id AS answer_id,
  sa.participant_id,
  sa.question_id,
  sa.answer,
  sa.manual_score,
  sa.graded_at,
  q.code AS question_code,
  q.question_text,
  s.nis, s.nama,
  COALESCE(r.room_name, '') AS room_name
FROM cbt_student_answers sa
JOIN cbt_questions q ON q.id = sa.question_id
JOIN cbt_exam_participants ep ON ep.id = sa.participant_id
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.session_id = $1
  AND q.question_type = 'essay'
  AND sa.manual_score IS NULL
ORDER BY q.code, s.nama;

