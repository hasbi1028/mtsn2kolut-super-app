-- name: ListCbtExamSessions :many
SELECT
  s.id, s.package_id, p.title AS package_title,
  s.class_id, c.name AS class_name, c.code AS class_code,
  s.title, s.scheduled_start, s.scheduled_end, s.status,
  s.created_at, s.updated_at,
  COUNT(ep.id)::int AS participant_count
FROM cbt_exam_sessions s
JOIN cbt_packages p ON p.id = s.package_id
JOIN school_classes c ON c.id = s.class_id
LEFT JOIN cbt_exam_participants ep ON ep.session_id = s.id
GROUP BY s.id, p.title, c.name, c.code
ORDER BY s.scheduled_start DESC;

-- name: GetCbtExamSession :one
SELECT
  s.id, s.package_id, p.title AS package_title, p.duration_minutes,
  s.class_id, c.name AS class_name, c.code AS class_code,
  s.title, s.scheduled_start, s.scheduled_end, s.status,
  s.created_at, s.updated_at
FROM cbt_exam_sessions s
JOIN cbt_packages p ON p.id = s.package_id
JOIN school_classes c ON c.id = s.class_id
WHERE s.id = $1;

-- name: CreateCbtExamSession :one
INSERT INTO cbt_exam_sessions (id, package_id, class_id, title, scheduled_start, scheduled_end, status)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateCbtExamSessionStatus :one
UPDATE cbt_exam_sessions
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCbtExamSession :exec
DELETE FROM cbt_exam_sessions WHERE id = $1 AND status = 'draft';

-- name: ListCbtExamParticipants :many
SELECT
  ep.id, ep.session_id, ep.student_id,
  s.nis, s.nama, s.gender,
  ep.token, ep.joined_at, ep.submitted_at, ep.score,
  ep.created_at
FROM cbt_exam_participants ep
JOIN students s ON s.id = ep.student_id
WHERE ep.session_id = $1
ORDER BY s.nama ASC;

-- name: EnrollClassToSession :exec
INSERT INTO cbt_exam_participants (session_id, student_id)
SELECT $1, s.id
FROM students s
WHERE s.class_id = $2 AND s.is_active = TRUE
ON CONFLICT (session_id, student_id) DO NOTHING;

-- name: UpsertStudentAnswer :exec
INSERT INTO cbt_student_answers (participant_id, question_id, answer, is_correct)
VALUES ($1, $2, $3, NULL)
ON CONFLICT (participant_id, question_id)
DO UPDATE SET answer = EXCLUDED.answer, is_correct = NULL, answered_at = NOW();

-- name: UpdateAnswerCorrectness :exec
UPDATE cbt_student_answers sa
SET is_correct = (sa.answer = q.answer_key)
FROM cbt_questions q
WHERE sa.question_id = q.id
  AND sa.participant_id IN (
    SELECT id FROM cbt_exam_participants WHERE session_id = $1
  );

-- name: UpdateParticipantScores :exec
UPDATE cbt_exam_participants ep
SET
  score = subq.pct,
  submitted_at = COALESCE(ep.submitted_at, NOW())
FROM (
  SELECT
    sa.participant_id,
    ROUND(
      SUM(CASE WHEN sa.is_correct THEN 1 ELSE 0 END)::numeric /
      NULLIF(COUNT(sa.id), 0) * 100, 2
    ) AS pct
  FROM cbt_student_answers sa
  WHERE sa.participant_id IN (
    SELECT ep2.id FROM cbt_exam_participants ep2 WHERE ep2.session_id = $1
  )
  GROUP BY sa.participant_id
) subq
WHERE ep.id = subq.participant_id;

-- name: GetSessionResults :many
SELECT
  ep.id AS participant_id,
  ep.student_id,
  s.nis, s.nama, s.gender,
  ep.submitted_at, ep.score,
  COUNT(sa.id)::int               AS total_answers,
  SUM(CASE WHEN sa.is_correct THEN 1 ELSE 0 END)::int AS correct_answers
FROM cbt_exam_participants ep
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_student_answers sa ON sa.participant_id = ep.id
WHERE ep.session_id = $1
GROUP BY ep.id, s.nis, s.nama, s.gender
ORDER BY ep.score DESC NULLS LAST, s.nama ASC;

-- name: GetParticipantAnswers :many
SELECT
  sa.id, sa.participant_id, sa.question_id,
  q.code AS question_code, q.question_text,
  q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
  q.answer_key,
  sa.answer, sa.is_correct, sa.answered_at
FROM cbt_student_answers sa
JOIN cbt_questions q ON q.id = sa.question_id
WHERE sa.participant_id = $1
ORDER BY q.code ASC;
