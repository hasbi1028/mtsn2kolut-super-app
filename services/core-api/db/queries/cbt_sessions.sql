-- name: ListCbtExamSessions :many
SELECT
  s.id, s.package_id, p.title AS package_title,
  s.class_id, s.event_id,
  COALESCE(c.name, '') AS class_name, COALESCE(c.code, '') AS class_code,
  s.scope_type, s.scope_ref, s.mix_policy, s.assignment_mode, s.allow_cross_grade, s.is_special_event,
  s.title, s.scheduled_start, s.scheduled_end, s.status,
  s.created_at, s.updated_at,
  COUNT(ep.id)::int AS participant_count
FROM cbt_exam_sessions s
JOIN cbt_packages p ON p.id = s.package_id
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN cbt_exam_participants ep ON ep.session_id = s.id
GROUP BY s.id, p.title, c.name, c.code
ORDER BY s.scheduled_start DESC;

-- name: GetCbtExamSession :one
SELECT
  s.id, s.package_id, p.title AS package_title, p.duration_minutes,
  s.class_id, s.event_id,
  COALESCE(c.name, '') AS class_name, COALESCE(c.code, '') AS class_code,
  s.scope_type, s.scope_ref, s.mix_policy, s.assignment_mode, s.allow_cross_grade, s.is_special_event,
  s.title, s.scheduled_start, s.scheduled_end, s.status,
  s.created_at, s.updated_at
FROM cbt_exam_sessions s
JOIN cbt_packages p ON p.id = s.package_id
LEFT JOIN school_classes c ON c.id = s.class_id
WHERE s.id = $1;

-- name: CreateCbtExamSession :one
INSERT INTO cbt_exam_sessions (
  package_id, class_id, event_id, scope_type, scope_ref, mix_policy, assignment_mode,
  allow_cross_grade, is_special_event, title, scheduled_start, scheduled_end, status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
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
  ep.token, ep.room_id, ep.seat_no, ep.joined_at, ep.submitted_at, ep.score,
  ep.app_switch_count, ep.screenshot_attempt, ep.suspicious_flag,
  ep.last_heartbeat, ep.created_at,
  COALESCE(r.room_name, '') AS room_name
FROM cbt_exam_participants ep
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.session_id = $1
ORDER BY s.nama ASC;

-- name: GetParticipantByToken :one
SELECT
  ep.id, ep.session_id, ep.student_id,
  ep.token, ep.room_id, ep.seat_no, ep.device_fingerprint, ep.question_order,
  ep.joined_at, ep.submitted_at, ep.score,
  ep.app_switch_count, ep.screenshot_attempt, ep.suspicious_flag,
  ep.last_heartbeat,
  s.nis, s.nama, s.gender,
  cs.status AS session_status,
  cs.title AS session_title,
  cs.scheduled_start, cs.scheduled_end,
  cs.package_id,
  p.title AS package_title,
  p.duration_minutes
FROM cbt_exam_participants ep
JOIN students s ON s.id = ep.student_id
JOIN cbt_exam_sessions cs ON cs.id = ep.session_id
JOIN cbt_packages p ON p.id = cs.package_id
WHERE ep.token = $1;

-- name: EnrollClassToSession :exec
INSERT INTO cbt_exam_participants (session_id, student_id, token)
SELECT $1, s.id, encode(gen_random_bytes(4), 'hex')
FROM students s
WHERE s.class_id = $2 AND s.is_active = TRUE
ON CONFLICT (session_id, student_id) DO NOTHING;

-- name: EnrollGradeToSession :exec
INSERT INTO cbt_exam_participants (session_id, student_id, token)
SELECT $1, s.id, encode(gen_random_bytes(4), 'hex')
FROM students s
JOIN school_classes c ON c.id = s.class_id
WHERE c.level = $2 AND s.is_active = TRUE
ON CONFLICT (session_id, student_id) DO NOTHING;

-- name: EnrollSchoolToSession :exec
INSERT INTO cbt_exam_participants (session_id, student_id, token)
SELECT $1, s.id, encode(gen_random_bytes(4), 'hex')
FROM students s
WHERE s.is_active = TRUE
ON CONFLICT (session_id, student_id) DO NOTHING;

-- name: GenerateTokensForSession :exec
UPDATE cbt_exam_participants
SET token = encode(gen_random_bytes(4), 'hex')
WHERE session_id = $1 AND (token = '' OR token IS NULL);

-- name: RegenerateParticipantToken :one
UPDATE cbt_exam_participants
SET token = encode(gen_random_bytes(4), 'hex')
WHERE id = $1
RETURNING id, token;

-- name: ResetParticipantRuntimeAccess :exec
UPDATE cbt_exam_participants
SET device_fingerprint = NULL,
    login_ip = NULL,
    last_heartbeat = NULL,
    question_order = NULL
WHERE id = $1;

-- name: AssignParticipantRoom :exec
UPDATE cbt_exam_participants
SET room_id = $2
WHERE id = $1;

-- name: ClearParticipantRooms :exec
UPDATE cbt_exam_participants
SET room_id = NULL, seat_no = NULL
WHERE session_id = $1;

-- name: AssignParticipantSeat :exec
UPDATE cbt_exam_participants
SET room_id = $2,
    seat_no = $3
WHERE id = $1;

-- name: UpdateParticipantQuestionOrder :exec
UPDATE cbt_exam_participants
SET question_order = $2
WHERE id = $1;

-- name: UpdateParticipantLogin :exec
UPDATE cbt_exam_participants
SET device_fingerprint = $2,
    login_ip           = $3,
    joined_at          = COALESCE(joined_at, NOW()),
    last_heartbeat     = NOW()
WHERE id = $1;

-- name: UpdateParticipantHeartbeat :exec
UPDATE cbt_exam_participants
SET last_heartbeat = NOW()
WHERE id = $1;

-- name: IncrementParticipantAppSwitch :exec
UPDATE cbt_exam_participants
SET app_switch_count = app_switch_count + 1
WHERE id = $1;

-- name: IncrementParticipantScreenshot :exec
UPDATE cbt_exam_participants
SET screenshot_attempt = screenshot_attempt + 1
WHERE id = $1;

-- name: SetParticipantSuspiciousFlag :exec
UPDATE cbt_exam_participants
SET suspicious_flag = $2
WHERE id = $1;

-- name: SubmitParticipantExam :one
UPDATE cbt_exam_participants
SET submitted_at = NOW()
WHERE id = $1 AND submitted_at IS NULL
RETURNING id, submitted_at;

-- name: InsertParticipantEvent :exec
INSERT INTO cbt_participant_events (participant_id, event_type, event_data)
VALUES ($1, $2, $3);

-- name: ListParticipantEvents :many
SELECT id, participant_id, event_type, event_data, created_at
FROM cbt_participant_events
WHERE participant_id = $1
ORDER BY created_at DESC
LIMIT 100;

-- name: ListSessionParticipantEvents :many
SELECT
  ev.id,
  ev.participant_id,
  ep.student_id,
  s.nis,
  s.nama,
  COALESCE(r.room_name, '') AS room_name,
  ev.event_type,
  ev.event_data,
  ev.created_at
FROM cbt_participant_events ev
JOIN cbt_exam_participants ep ON ep.id = ev.participant_id
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.session_id = $1
  AND (sqlc.arg(participant_id)::uuid IS NULL OR ev.participant_id = sqlc.arg(participant_id)::uuid)
ORDER BY ev.created_at DESC
LIMIT sqlc.arg(limit_count);

-- name: GetSessionProctoringStatus :many
SELECT
  ep.id AS participant_id,
  ep.student_id,
  s.nis, s.nama,
  ep.token,
  COALESCE(r.room_name, '') AS room_name,
  ep.seat_no,
  ep.submitted_at,
  ep.last_heartbeat,
  ep.app_switch_count,
  ep.screenshot_attempt,
  ep.suspicious_flag,
  COUNT(sa.id)::int AS answered_count,
  ep.score
FROM cbt_exam_participants ep
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
LEFT JOIN cbt_student_answers sa ON sa.participant_id = ep.id
WHERE ep.session_id = $1
GROUP BY ep.id, s.nis, s.nama, r.room_name
ORDER BY s.nama ASC;

-- name: ListParticipantsByRoom :many
SELECT
  ep.id, ep.student_id, ep.token, ep.room_id, ep.seat_no,
  s.nis, s.nama, s.gender,
  COALESCE(r.room_name, '') AS room_name
FROM cbt_exam_participants ep
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.session_id = $1
ORDER BY r.room_name ASC NULLS LAST, ep.seat_no ASC NULLS LAST, s.nama ASC;

-- name: ListStudentExamSessions :many
SELECT
  ep.id AS participant_id,
  ep.session_id,
  ep.token,
  ep.room_id,
  ep.seat_no,
  ep.joined_at,
  ep.submitted_at,
  ep.score,
  s.title AS session_title,
  s.status AS session_status,
  s.scheduled_start,
  s.scheduled_end,
  p.title AS package_title,
  p.duration_minutes,
  COALESCE(r.room_name, '') AS room_name
FROM cbt_exam_participants ep
JOIN cbt_exam_sessions s ON s.id = ep.session_id
JOIN cbt_packages p ON p.id = s.package_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.student_id = $1
ORDER BY s.scheduled_start DESC;

-- name: GradeStudentEssay :exec
UPDATE cbt_student_answers
SET manual_score = $2,
    graded_by    = $3,
    graded_at    = NOW(),
    is_correct   = NULL
WHERE id = $1;

-- name: UpsertStudentAnswer :exec
INSERT INTO cbt_student_answers (participant_id, question_id, answer, is_correct)
VALUES ($1, $2, $3, NULL)
ON CONFLICT (participant_id, question_id)
DO UPDATE SET answer = EXCLUDED.answer, is_correct = NULL, answered_at = NOW();

-- name: UpdateAnswerCorrectness :exec
UPDATE cbt_student_answers sa
SET is_correct = CASE
  -- essay: skip, scored manually
  WHEN q.question_type = 'essay' THEN NULL
  -- multiple_answer: answer is comma-separated labels, must match answer_key exactly after sorting
  WHEN q.question_type = 'multiple_answer' THEN
    (array_to_string(ARRAY(SELECT btrim(label) FROM unnest(string_to_array(sa.answer, ',')) AS key(label) WHERE btrim(label) <> '' ORDER BY 1), ',') =
     array_to_string(ARRAY(SELECT btrim(label) FROM unnest(string_to_array(q.answer_key, ',')) AS key(label) WHERE btrim(label) <> '' ORDER BY 1), ','))
  -- matching: answer is semicolon-separated left=right pairs, all pairs must match.
  WHEN q.question_type = 'matching' THEN
    (array_to_string(ARRAY(SELECT upper(btrim(pair)) FROM unnest(string_to_array(sa.answer, ';')) AS key(pair) WHERE btrim(pair) <> '' ORDER BY 1), ';') =
     array_to_string(ARRAY(SELECT upper(btrim(pair)) FROM unnest(string_to_array(q.answer_key, ';')) AS key(pair) WHERE btrim(pair) <> '' ORDER BY 1), ';'))
  -- short_answer: answer_key may contain accepted aliases separated by "|";
  -- normalize case, repeated whitespace, and non-breaking spaces before matching.
  WHEN q.question_type = 'short_answer' THEN EXISTS (
    SELECT 1
    FROM unnest(string_to_array(q.answer_key, '|')) AS accepted(answer)
    WHERE lower(regexp_replace(btrim(replace(accepted.answer, chr(160), ' ')), '[[:space:]]+', ' ', 'g')) =
          lower(regexp_replace(btrim(replace(sa.answer, chr(160), ' ')), '[[:space:]]+', ' ', 'g'))
  )
  -- all others: exact string match
  ELSE (sa.answer = q.answer_key)
END
FROM cbt_questions q
WHERE sa.question_id = q.id
  AND q.question_type <> 'essay'
  AND sa.manual_score IS NULL
  AND sa.participant_id IN (
    SELECT id FROM cbt_exam_participants WHERE session_id = $1
  );

-- name: UpdateParticipantAnswerCorrectness :exec
UPDATE cbt_student_answers sa
SET is_correct = CASE
  WHEN q.question_type = 'essay' THEN NULL
  WHEN q.question_type = 'multiple_answer' THEN
    (array_to_string(ARRAY(SELECT btrim(label) FROM unnest(string_to_array(sa.answer, ',')) AS key(label) WHERE btrim(label) <> '' ORDER BY 1), ',') =
     array_to_string(ARRAY(SELECT btrim(label) FROM unnest(string_to_array(q.answer_key, ',')) AS key(label) WHERE btrim(label) <> '' ORDER BY 1), ','))
  WHEN q.question_type = 'matching' THEN
    (array_to_string(ARRAY(SELECT upper(btrim(pair)) FROM unnest(string_to_array(sa.answer, ';')) AS key(pair) WHERE btrim(pair) <> '' ORDER BY 1), ';') =
     array_to_string(ARRAY(SELECT upper(btrim(pair)) FROM unnest(string_to_array(q.answer_key, ';')) AS key(pair) WHERE btrim(pair) <> '' ORDER BY 1), ';'))
  WHEN q.question_type = 'short_answer' THEN EXISTS (
    SELECT 1
    FROM unnest(string_to_array(q.answer_key, '|')) AS accepted(answer)
    WHERE lower(regexp_replace(btrim(replace(accepted.answer, chr(160), ' ')), '[[:space:]]+', ' ', 'g')) =
          lower(regexp_replace(btrim(replace(sa.answer, chr(160), ' ')), '[[:space:]]+', ' ', 'g'))
  )
  ELSE (sa.answer = q.answer_key)
END
FROM cbt_questions q
WHERE sa.question_id = q.id
  AND q.question_type <> 'essay'
  AND sa.manual_score IS NULL
  AND sa.participant_id = $1;

-- name: UpdateParticipantScores :exec
WITH score_parts AS (
  SELECT
    ep.id AS participant_id,
    COALESCE(SUM(
      CASE
        WHEN q.question_type = 'essay' AND sa.manual_score IS NOT NULL THEN (sa.manual_score / 100) * pq.points
        WHEN q.question_type <> 'essay' AND sa.is_correct IS TRUE THEN pq.points
        ELSE 0
      END
    ), 0)::numeric AS earned_points,
    COALESCE(SUM(pq.points), 0)::numeric AS total_points
  FROM cbt_exam_participants ep
  JOIN cbt_exam_sessions ses ON ses.id = ep.session_id
  JOIN cbt_package_questions pq ON pq.package_id = ses.package_id
  JOIN cbt_questions q ON q.id = pq.question_id
  LEFT JOIN cbt_student_answers sa ON sa.participant_id = ep.id AND sa.question_id = pq.question_id
  WHERE ep.session_id = $1
  GROUP BY ep.id
)
UPDATE cbt_exam_participants ep
SET score = CASE
      WHEN score_parts.total_points > 0 THEN ROUND((score_parts.earned_points / score_parts.total_points) * 100, 2)
      ELSE 0
    END,
  submitted_at = COALESCE(ep.submitted_at, NOW())
FROM score_parts
WHERE ep.id = score_parts.participant_id;

-- name: ForceSubmitParticipant :one
WITH score_parts AS (
  SELECT
    ep.id AS participant_id,
    COALESCE(SUM(
      CASE
        WHEN q.question_type = 'essay' AND sa.manual_score IS NOT NULL THEN (sa.manual_score / 100) * pq.points
        WHEN q.question_type <> 'essay' AND sa.is_correct IS TRUE THEN pq.points
        ELSE 0
      END
    ), 0)::numeric AS earned_points,
    COALESCE(SUM(pq.points), 0)::numeric AS total_points
  FROM cbt_exam_participants ep
  JOIN cbt_exam_sessions ses ON ses.id = ep.session_id
  JOIN cbt_package_questions pq ON pq.package_id = ses.package_id
  JOIN cbt_questions q ON q.id = pq.question_id
  LEFT JOIN cbt_student_answers sa ON sa.participant_id = ep.id AND sa.question_id = pq.question_id
  WHERE ep.session_id = $1 AND ep.id = $2
  GROUP BY ep.id
)
UPDATE cbt_exam_participants ep
SET score = CASE
      WHEN score_parts.total_points > 0 THEN ROUND((score_parts.earned_points / score_parts.total_points) * 100, 2)
      ELSE 0
    END,
    submitted_at = COALESCE(ep.submitted_at, NOW())
FROM score_parts
WHERE ep.id = score_parts.participant_id
RETURNING ep.id, ep.submitted_at, ep.score;

-- name: GetSessionResults :many
SELECT
  ep.id AS participant_id,
  ep.student_id,
  s.nis, s.nama, s.gender,
  ep.submitted_at, ep.score,
  ep.room_id, ep.seat_no,
  COALESCE(r.room_name, '') AS room_name,
  COUNT(sa.id)::int                                    AS total_answers,
  SUM(CASE WHEN sa.is_correct THEN 1 ELSE 0 END)::int AS correct_answers
FROM cbt_exam_participants ep
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
LEFT JOIN cbt_student_answers sa ON sa.participant_id = ep.id
WHERE ep.session_id = $1
GROUP BY ep.id, s.nis, s.nama, s.gender, ep.room_id, ep.seat_no, r.room_name
ORDER BY ep.score DESC NULLS LAST, s.nama ASC;

-- name: ListCbtExamSessionsByTeacher :many
SELECT
  s.id, s.package_id, p.title AS package_title,
  s.class_id, s.event_id,
  COALESCE(c.name, '') AS class_name, COALESCE(c.code, '') AS class_code,
  s.scope_type, s.scope_ref, s.mix_policy, s.assignment_mode, s.allow_cross_grade, s.is_special_event,
  s.title, s.scheduled_start, s.scheduled_end, s.status,
  s.created_at, s.updated_at,
  COUNT(ep.id)::int AS participant_count
FROM cbt_exam_sessions s
JOIN cbt_packages p ON p.id = s.package_id
JOIN class_subject_assignments csa ON csa.subject_id = p.subject_id
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN cbt_exam_participants ep ON ep.session_id = s.id
WHERE csa.teacher_employee_id = $1
GROUP BY s.id, p.title, c.name, c.code
ORDER BY s.scheduled_start DESC;

-- name: GetSessionResultsByTeacher :many
SELECT
  ep.id AS participant_id,
  ep.student_id,
  s.nis, s.nama, s.gender,
  ep.submitted_at, ep.score,
  COUNT(sa.id)::int AS total_answers,
  SUM(CASE WHEN sa.is_correct THEN 1 ELSE 0 END)::int AS correct_answers
FROM cbt_exam_participants ep
JOIN students s ON s.id = ep.student_id
JOIN cbt_exam_sessions ses ON ses.id = ep.session_id
JOIN cbt_packages pkg ON pkg.id = ses.package_id
JOIN class_subject_assignments csa ON csa.subject_id = pkg.subject_id
LEFT JOIN cbt_student_answers sa ON sa.participant_id = ep.id
WHERE ep.session_id = $1 AND csa.teacher_employee_id = $2
GROUP BY ep.id, s.nis, s.nama, s.gender
ORDER BY ep.score DESC NULLS LAST, s.nama ASC;

-- name: GetSessionTeacherAccess :one
SELECT EXISTS(
  SELECT 1 FROM cbt_exam_sessions s
  JOIN cbt_packages p ON p.id = s.package_id
  JOIN class_subject_assignments csa ON csa.subject_id = p.subject_id
  WHERE s.id = $1 AND csa.teacher_employee_id = $2
) AS has_access;

-- name: HasSessionParticipant :one
SELECT EXISTS(
  SELECT 1
  FROM cbt_exam_participants
  WHERE session_id = $1 AND id = $2
) AS has_participant;

-- name: HasSessionRoom :one
SELECT EXISTS(
  SELECT 1
  FROM cbt_exam_rooms
  WHERE session_id = $1 AND id = $2
) AS has_room;

-- name: HasSessionAnswer :one
SELECT EXISTS(
  SELECT 1
  FROM cbt_student_answers sa
  JOIN cbt_exam_participants ep ON ep.id = sa.participant_id
  WHERE ep.session_id = $1 AND sa.id = $2
) AS has_answer;

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
