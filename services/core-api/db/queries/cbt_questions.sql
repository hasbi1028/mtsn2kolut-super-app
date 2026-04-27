-- name: ListCbtQuestions :many
SELECT q.id, q.subject_id, s.name AS subject_name, s.code AS subject_code,
       q.code, q.question_text, q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
       q.answer_key, q.explanation, q.difficulty, q.status, q.created_at, q.updated_at
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
ORDER BY q.created_at DESC;

-- name: GetCbtQuestion :one
SELECT id, subject_id, code, question_text, option_a, option_b, option_c, option_d, option_e,
       answer_key, explanation, difficulty, status, created_at, updated_at
FROM cbt_questions
WHERE id = $1;

-- name: CreateCbtQuestion :one
INSERT INTO cbt_questions (
  id, subject_id, code, question_text, option_a, option_b, option_c, option_d, option_e,
  answer_key, explanation, difficulty, status
)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: DeleteCbtQuestion :exec
DELETE FROM cbt_questions WHERE id = $1;
