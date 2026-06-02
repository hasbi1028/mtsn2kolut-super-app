-- name: ListAssessmentExams :many
SELECT
  e.*,
  COUNT(DISTINCT s.id)::bigint AS session_count,
  COUNT(DISTINCT r.id)::bigint AS room_count,
  COUNT(DISTINCT p.id)::bigint AS participant_count,
  COUNT(DISTINCT ac.id)::bigint AS card_count
FROM assessment_exams e
LEFT JOIN assessment_sessions s ON s.exam_id = e.id
LEFT JOIN assessment_rooms r ON r.session_id = s.id
LEFT JOIN assessment_participants p ON p.session_id = s.id
LEFT JOIN assessment_access_cards ac ON ac.session_id = s.id AND ac.card_type = 'participant'
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
LEFT JOIN assessment_access_cards c ON c.session_id = s.id AND c.card_type = 'participant'
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

-- name: UpsertAssessmentRoom :one
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
JOIN assessment_access_cards c ON c.session_id = s.id AND c.card_type = 'participant'
WHERE s.exam_id = sqlc.arg(exam_id);

-- name: ListAssessmentCandidateStudentsByClassIDs :many
SELECT
  s.id AS student_id,
  s.nama AS student_name,
  s.nis,
  s.nisn,
  s.class_id,
  COALESCE(c.code, '') AS class_code,
  COALESCE(c.name, '') AS class_name,
  CASE UPPER(NULLIF(btrim(c.level::text), ''))
    WHEN '7' THEN 7
    WHEN 'VII' THEN 7
    WHEN '8' THEN 8
    WHEN 'VIII' THEN 8
    WHEN '9' THEN 9
    WHEN 'IX' THEN 9
    ELSE 0
  END::int AS grade_level
FROM students s
LEFT JOIN school_classes c ON c.id = s.class_id
WHERE s.is_active = TRUE
  AND s.class_id = ANY(sqlc.arg(class_ids)::uuid[])
ORDER BY CASE UPPER(NULLIF(btrim(c.level::text), ''))
    WHEN '7' THEN 7
    WHEN 'VII' THEN 7
    WHEN '8' THEN 8
    WHEN 'VIII' THEN 8
    WHEN '9' THEN 9
    WHEN 'IX' THEN 9
    ELSE 0
  END, COALESCE(c.code, ''), s.nama;

-- name: UpsertAssessmentParticipant :one
INSERT INTO assessment_participants (
  session_id,
  student_id,
  status
) VALUES (
  sqlc.arg(session_id),
  sqlc.arg(student_id),
  'registered'
)
ON CONFLICT (session_id, student_id) DO UPDATE
SET updated_at = now()
RETURNING *;

-- name: DeleteAssessmentParticipantsOutsideClassIDs :execrows
DELETE FROM assessment_participants p
USING students s
WHERE p.session_id = sqlc.arg(session_id)
  AND s.id = p.student_id
  AND (
    s.class_id IS NULL
    OR NOT (s.class_id = ANY(sqlc.arg(class_ids)::uuid[]))
  );

-- name: ListAssessmentParticipantsForAssignment :many
SELECT
  p.id AS participant_id,
  p.session_id,
  p.room_id,
  p.student_id,
  p.status,
  COALESCE(p.seat_no, 0)::int AS seat_no,
  s.nama AS student_name,
  s.class_id,
  COALESCE(c.code, '') AS class_code,
  COALESCE(c.name, '') AS class_name,
  CASE UPPER(NULLIF(btrim(c.level::text), ''))
    WHEN '7' THEN 7
    WHEN 'VII' THEN 7
    WHEN '8' THEN 8
    WHEN 'VIII' THEN 8
    WHEN '9' THEN 9
    WHEN 'IX' THEN 9
    ELSE 0
  END::int AS grade_level
FROM assessment_participants p
JOIN students s ON s.id = p.student_id
LEFT JOIN school_classes c ON c.id = s.class_id
WHERE p.session_id = sqlc.arg(session_id)
ORDER BY CASE UPPER(NULLIF(btrim(c.level::text), ''))
    WHEN '7' THEN 7
    WHEN 'VII' THEN 7
    WHEN '8' THEN 8
    WHEN 'VIII' THEN 8
    WHEN '9' THEN 9
    WHEN 'IX' THEN 9
    ELSE 0
  END, COALESCE(c.code, ''), s.nama, p.id;

-- name: ClearAssessmentParticipantRooms :exec
UPDATE assessment_participants
SET room_id = NULL,
    seat_no = NULL,
    updated_at = now()
WHERE session_id = sqlc.arg(session_id);

-- name: AssignAssessmentParticipantRoom :exec
UPDATE assessment_participants
SET room_id = sqlc.arg(room_id),
    seat_no = sqlc.arg(seat_no),
    updated_at = now()
WHERE id = sqlc.arg(participant_id)
  AND session_id = sqlc.arg(session_id);


-- name: ListAssessmentParticipantPlacementsByExam :many
SELECT
  p.id AS participant_id,
  p.session_id,
  p.room_id,
  p.student_id,
  p.status,
  COALESCE(p.seat_no, 0)::int AS seat_no,
  s.nama AS student_name,
  s.nis,
  s.nisn,
  s.class_id,
  COALESCE(c.code, '') AS class_code,
  COALESCE(c.name, '') AS class_name,
  CASE UPPER(NULLIF(btrim(c.level::text), ''))
    WHEN '7' THEN 7
    WHEN 'VII' THEN 7
    WHEN '8' THEN 8
    WHEN 'VIII' THEN 8
    WHEN '9' THEN 9
    WHEN 'IX' THEN 9
    ELSE 0
  END::int AS grade_level,
  r.id AS room_id_actual,
  COALESCE(r.code, '') AS room_code,
  COALESCE(r.name, '') AS room_name,
  COALESCE(r.capacity, 0)::int AS room_capacity
FROM assessment_participants p
JOIN assessment_sessions sess ON sess.id = p.session_id
JOIN students s ON s.id = p.student_id
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN assessment_rooms r ON r.id = p.room_id
WHERE sess.exam_id = sqlc.arg(exam_id)
ORDER BY COALESCE(r.code, 'ZZZ'), COALESCE(p.seat_no, 9999),
  CASE UPPER(NULLIF(btrim(c.level::text), ''))
    WHEN '7' THEN 7
    WHEN 'VII' THEN 7
    WHEN '8' THEN 8
    WHEN 'VIII' THEN 8
    WHEN '9' THEN 9
    WHEN 'IX' THEN 9
    ELSE 0
  END, COALESCE(c.code, ''), s.nama, p.id;

-- name: GetAssessmentRoomByIDAndExam :one
SELECT r.id, r.session_id, r.code, r.name, r.capacity
FROM assessment_rooms r
JOIN assessment_sessions s ON s.id = r.session_id
WHERE s.exam_id = sqlc.arg(exam_id)
  AND r.id = sqlc.arg(room_id)
LIMIT 1;

-- name: MoveAssessmentParticipantSeat :one
WITH updated AS (
  UPDATE assessment_participants p
  SET room_id = sqlc.arg(room_id),
      seat_no = sqlc.arg(seat_no),
      updated_at = now()
  FROM assessment_sessions sess
  WHERE p.id = sqlc.arg(participant_id)
    AND p.session_id = sess.id
    AND sess.exam_id = sqlc.arg(exam_id)
  RETURNING p.id AS participant_id
)
SELECT
  p.id AS participant_id,
  p.session_id,
  p.room_id,
  p.student_id,
  p.status,
  COALESCE(p.seat_no, 0)::int AS seat_no,
  s.nama AS student_name,
  s.nis,
  s.nisn,
  s.class_id,
  COALESCE(c.code, '') AS class_code,
  COALESCE(c.name, '') AS class_name,
  CASE UPPER(NULLIF(btrim(c.level::text), ''))
    WHEN '7' THEN 7
    WHEN 'VII' THEN 7
    WHEN '8' THEN 8
    WHEN 'VIII' THEN 8
    WHEN '9' THEN 9
    WHEN 'IX' THEN 9
    ELSE 0
  END::int AS grade_level,
  r.id AS room_id_actual,
  COALESCE(r.code, '') AS room_code,
  COALESCE(r.name, '') AS room_name,
  COALESCE(r.capacity, 0)::int AS room_capacity
FROM updated u
JOIN assessment_participants p ON p.id = u.participant_id
JOIN students s ON s.id = p.student_id
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN assessment_rooms r ON r.id = p.room_id;


-- name: ListAssessmentParticipantCardTargetsByExam :many
SELECT
  p.id AS participant_id,
  p.session_id,
  p.room_id,
  p.student_id,
  p.status,
  COALESCE(p.seat_no, 0)::int AS seat_no,
  s.nama AS student_name,
  s.nis,
  s.nisn,
  s.class_id,
  COALESCE(c.code, '') AS class_code,
  COALESCE(c.name, '') AS class_name,
  CASE UPPER(NULLIF(btrim(c.level::text), ''))
    WHEN '7' THEN 7
    WHEN 'VII' THEN 7
    WHEN '8' THEN 8
    WHEN 'VIII' THEN 8
    WHEN '9' THEN 9
    WHEN 'IX' THEN 9
    ELSE 0
  END::int AS grade_level,
  COALESCE(r.code, '') AS room_code,
  COALESCE(r.name, '') AS room_name,
  COALESCE(r.capacity, 0)::int AS room_capacity,
  ac.id AS card_id,
  COALESCE(ac.status, 'not_issued') AS card_status,
  COALESCE(ac.failed_attempts, 0)::int AS failed_attempts,
  ac.expires_at AS card_expires_at,
  ac.created_at AS card_created_at
FROM assessment_participants p
JOIN assessment_sessions sess ON sess.id = p.session_id
JOIN students s ON s.id = p.student_id
LEFT JOIN school_classes c ON c.id = s.class_id
LEFT JOIN assessment_rooms r ON r.id = p.room_id
LEFT JOIN assessment_access_cards ac ON ac.participant_id = p.id AND ac.card_type = 'participant'
WHERE sess.exam_id = sqlc.arg(exam_id)
ORDER BY COALESCE(r.code, 'ZZZ'), COALESCE(p.seat_no, 9999),
  CASE UPPER(NULLIF(btrim(c.level::text), ''))
    WHEN '7' THEN 7
    WHEN 'VII' THEN 7
    WHEN '8' THEN 8
    WHEN 'VIII' THEN 8
    WHEN '9' THEN 9
    WHEN 'IX' THEN 9
    ELSE 0
  END, COALESCE(c.code, ''), s.nama, p.id
FOR UPDATE OF p;

-- name: CreateAssessmentParticipantAccessCard :one
INSERT INTO assessment_access_cards (
  card_type,
  session_id,
  participant_id,
  token_hash,
  pin_hash,
  status,
  expires_at
) VALUES (
  'participant',
  sqlc.arg(session_id),
  sqlc.arg(participant_id),
  sqlc.arg(token_hash),
  sqlc.arg(pin_hash),
  'active',
  sqlc.narg(expires_at)
)
ON CONFLICT (participant_id) DO UPDATE
SET token_hash = EXCLUDED.token_hash,
    pin_hash = EXCLUDED.pin_hash,
    status = 'active',
    failed_attempts = 0,
    expires_at = EXCLUDED.expires_at,
    updated_at = now()
RETURNING *;
