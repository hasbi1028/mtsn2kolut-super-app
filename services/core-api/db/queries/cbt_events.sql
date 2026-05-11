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

-- name: DeleteCbtExamEvent :execrows
DELETE FROM cbt_exam_events WHERE id = $1 AND status = 'draft';

-- name: GetCbtEventOverviewSummary :one
SELECT
  e.id,
  e.title,
  e.exam_type,
  e.scope,
  e.target_levels,
  e.status,
  e.academic_year_id,
  COALESCE(ay.name, '') AS academic_year_name,
  e.created_at,
  e.updated_at,
  COALESCE(member_summary.member_count, 0)::int AS member_count,
  COALESCE(member_summary.panitia_count, 0)::int AS panitia_count,
  COALESCE(member_summary.author_count, 0)::int AS author_count,
  COALESCE(member_summary.reviewer_count, 0)::int AS reviewer_count,
  COALESCE(member_summary.proctor_count, 0)::int AS proctor_count,
  COALESCE(target_summary.target_subject_count, 0)::int AS target_subject_count,
  COALESCE(target_summary.target_question_count, 0)::int AS target_question_count,
  COALESCE(question_summary.total_questions, 0)::int AS total_questions,
  COALESCE(question_summary.published_questions, 0)::int AS published_questions,
  COALESCE(question_summary.review_questions, 0)::int AS review_questions,
  COALESCE(question_summary.approved_questions, 0)::int AS approved_questions,
  COALESCE(package_summary.package_count, 0)::int AS package_count,
  COALESCE(package_summary.active_package_count, 0)::int AS active_package_count,
  COALESCE(package_summary.empty_package_count, 0)::int AS empty_package_count,
  COALESCE(session_summary.session_count, 0)::int AS session_count,
  COALESCE(session_summary.draft_session_count, 0)::int AS draft_session_count,
  COALESCE(session_summary.scheduled_session_count, 0)::int AS scheduled_session_count,
  COALESCE(session_summary.active_session_count, 0)::int AS active_session_count,
  COALESCE(session_summary.finished_session_count, 0)::int AS finished_session_count,
  COALESCE(ops_summary.room_count, 0)::int AS room_count,
  COALESCE(ops_summary.rooms_without_proctor, 0)::int AS rooms_without_proctor,
  COALESCE(ops_summary.total_capacity, 0)::int AS total_capacity,
  COALESCE(participant_summary.participant_count, 0)::int AS participant_count,
  COALESCE(participant_summary.assigned_participant_count, 0)::int AS assigned_participant_count,
  COALESCE(participant_summary.unassigned_participant_count, 0)::int AS unassigned_participant_count,
  COALESCE(participant_summary.missing_seat_count, 0)::int AS missing_seat_count,
  COALESCE(participant_summary.token_ready_count, 0)::int AS token_ready_count,
  COALESCE(participant_summary.joined_count, 0)::int AS joined_count,
  COALESCE(participant_summary.submitted_count, 0)::int AS submitted_count,
  COALESCE(participant_summary.scored_count, 0)::int AS scored_count
FROM cbt_exam_events e
LEFT JOIN academic_years ay ON ay.id = e.academic_year_id
LEFT JOIN LATERAL (
  SELECT
    COUNT(*)::int AS member_count,
    COUNT(*) FILTER (WHERE role = 'panitia')::int AS panitia_count,
    COUNT(*) FILTER (WHERE role = 'pembuat_soal')::int AS author_count,
    COUNT(*) FILTER (WHERE role = 'reviewer')::int AS reviewer_count,
    COUNT(*) FILTER (WHERE role IN ('proktor', 'pengawas'))::int AS proctor_count
  FROM cbt_event_members m
  WHERE m.event_id = e.id
) member_summary ON TRUE
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS target_subject_count, COALESCE(SUM(target_questions), 0)::int AS target_question_count
  FROM cbt_event_subject_targets t
  WHERE t.event_id = e.id
) target_summary ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COUNT(*)::int AS total_questions,
    COUNT(*) FILTER (WHERE q.status = 'published')::int AS published_questions,
    COUNT(*) FILTER (WHERE q.workflow_status = 'review')::int AS review_questions,
    COUNT(*) FILTER (WHERE q.workflow_status = 'approved')::int AS approved_questions
  FROM cbt_questions q
  WHERE (q.event_id = e.id OR (q.event_id IS NULL AND q.status = 'published'))
    AND EXISTS (
      SELECT 1
      FROM cbt_event_subject_targets t
      WHERE t.event_id = e.id AND t.subject_id = q.subject_id
    )
) question_summary ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COUNT(*)::int AS package_count,
    COUNT(*) FILTER (WHERE p.is_active)::int AS active_package_count,
    COUNT(*) FILTER (WHERE COALESCE(pq.question_count, 0) = 0)::int AS empty_package_count
  FROM cbt_packages p
  LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS question_count
    FROM cbt_package_questions pq
    WHERE pq.package_id = p.id
  ) pq ON TRUE
  WHERE p.event_id = e.id
) package_summary ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COUNT(*)::int AS session_count,
    COUNT(*) FILTER (WHERE status = 'draft')::int AS draft_session_count,
    COUNT(*) FILTER (WHERE status = 'scheduled')::int AS scheduled_session_count,
    COUNT(*) FILTER (WHERE status = 'active')::int AS active_session_count,
    COUNT(*) FILTER (WHERE status = 'finished')::int AS finished_session_count
  FROM cbt_exam_sessions s
  WHERE s.event_id = e.id
) session_summary ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COUNT(r.id)::int AS room_count,
    COALESCE(SUM(r.capacity), 0)::int AS total_capacity,
    COUNT(r.id) FILTER (WHERE COALESCE(pr.proctor_count, 0) = 0)::int AS rooms_without_proctor
  FROM cbt_exam_sessions s
  JOIN cbt_exam_rooms r ON r.session_id = s.id
  LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS proctor_count
    FROM cbt_room_proctors rp
    WHERE rp.exam_room_id = r.id
  ) pr ON TRUE
  WHERE s.event_id = e.id
) ops_summary ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COUNT(ep.id)::int AS participant_count,
    COUNT(ep.id) FILTER (WHERE ep.room_id IS NOT NULL)::int AS assigned_participant_count,
    COUNT(ep.id) FILTER (WHERE ep.room_id IS NULL)::int AS unassigned_participant_count,
    COUNT(ep.id) FILTER (WHERE ep.room_id IS NOT NULL AND ep.seat_no IS NULL)::int AS missing_seat_count,
    COUNT(ep.id) FILTER (WHERE ep.token IS NOT NULL AND ep.token ~ '^[0-9a-f]{32}$')::int AS token_ready_count,
    COUNT(ep.id) FILTER (WHERE ep.joined_at IS NOT NULL)::int AS joined_count,
    COUNT(ep.id) FILTER (WHERE ep.submitted_at IS NOT NULL)::int AS submitted_count,
    COUNT(ep.id) FILTER (WHERE ep.score IS NOT NULL)::int AS scored_count
  FROM cbt_exam_sessions s
  JOIN cbt_exam_participants ep ON ep.session_id = s.id
  WHERE s.event_id = e.id
) participant_summary ON TRUE
WHERE e.id = $1;

-- name: ListCbtEventSubjectMatrix :many
SELECT
  subjects.subject_id,
  subjects.subject_name,
  subjects.subject_code,
  COALESCE(targets.target_questions, 0)::int AS target_questions,
  COALESCE(members.author_count, 0)::int AS author_count,
  COALESCE(members.reviewer_count, 0)::int AS reviewer_count,
  COALESCE(questions.total_questions, 0)::int AS total_questions,
  COALESCE(questions.published_questions, 0)::int AS published_questions,
  COALESCE(packages.package_count, 0)::int AS package_count,
  COALESCE(packages.active_package_count, 0)::int AS active_package_count,
  COALESCE(sessions.session_count, 0)::int AS session_count,
  GREATEST(COALESCE(targets.target_questions, 0)::int - COALESCE(questions.published_questions, 0)::int, 0)::int AS shortage_count,
  (COALESCE(targets.target_questions, 0)::int > 0 AND COALESCE(questions.published_questions, 0)::int >= COALESCE(targets.target_questions, 0)::int) AS authoring_ready,
  (COALESCE(packages.active_package_count, 0)::int > 0) AS package_ready,
  (COALESCE(sessions.session_count, 0)::int > 0) AS session_ready
FROM (
  SELECT t.event_id, t.subject_id, s.name AS subject_name, s.code AS subject_code
  FROM cbt_event_subject_targets t
  JOIN subjects s ON s.id = t.subject_id
  WHERE t.event_id = $1
  UNION
  SELECT q.event_id, q.subject_id, s.name AS subject_name, s.code AS subject_code
  FROM cbt_questions q
  JOIN subjects s ON s.id = q.subject_id
  WHERE q.event_id = $1
  UNION
  SELECT p.event_id, p.subject_id, s.name AS subject_name, s.code AS subject_code
  FROM cbt_packages p
  JOIN subjects s ON s.id = p.subject_id
  WHERE p.event_id = $1
) subjects
LEFT JOIN cbt_event_subject_targets targets ON targets.event_id = subjects.event_id AND targets.subject_id = subjects.subject_id
LEFT JOIN LATERAL (
  SELECT
    COUNT(*) FILTER (WHERE role = 'pembuat_soal')::int AS author_count,
    COUNT(*) FILTER (WHERE role = 'reviewer')::int AS reviewer_count
  FROM cbt_event_members m
  WHERE m.event_id = subjects.event_id
    AND (m.subject_id IS NULL OR m.subject_id = subjects.subject_id)
) members ON TRUE
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS total_questions, COUNT(*) FILTER (WHERE status = 'published')::int AS published_questions
  FROM cbt_questions q
  WHERE q.subject_id = subjects.subject_id
    AND (
      q.event_id = subjects.event_id
      OR (q.event_id IS NULL AND q.status = 'published')
    )
) questions ON TRUE
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS package_count, COUNT(*) FILTER (WHERE is_active)::int AS active_package_count
  FROM cbt_packages p
  WHERE p.event_id = subjects.event_id AND p.subject_id = subjects.subject_id
) packages ON TRUE
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS session_count
  FROM cbt_exam_sessions ses
  JOIN cbt_packages p ON p.id = ses.package_id
  WHERE ses.event_id = subjects.event_id AND p.subject_id = subjects.subject_id
) sessions ON TRUE
ORDER BY subjects.subject_name ASC, subjects.subject_code ASC;

-- name: GetCbtEventQuestionRequirements :one
SELECT
  COALESCE(r.id, '00000000-0000-0000-0000-000000000000'::uuid) AS id,
  e.id AS event_id,
  COALESCE(r.scope_mode, 'per_rombel')::text AS scope_mode,
  COALESCE(r.target_pg, 20)::int AS target_pg,
  COALESCE(r.target_essay, 5)::int AS target_essay,
  COALESCE(r.status_filter, 'published_only')::text AS status_filter,
  COALESCE(r.created_at, e.created_at) AS created_at,
  COALESCE(r.updated_at, e.updated_at) AS updated_at
FROM cbt_exam_events e
LEFT JOIN cbt_event_question_requirements r ON r.event_id = e.id
  AND r.level IS NULL
  AND r.class_id IS NULL
  AND r.subject_id IS NULL
WHERE e.id = $1
ORDER BY r.updated_at DESC NULLS LAST
LIMIT 1;

-- name: UpsertCbtEventQuestionRequirements :one
INSERT INTO cbt_event_question_requirements (event_id, scope_mode, target_pg, target_essay, status_filter)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (event_id) WHERE level IS NULL AND class_id IS NULL AND subject_id IS NULL
DO UPDATE SET
  scope_mode = EXCLUDED.scope_mode,
  target_pg = EXCLUDED.target_pg,
  target_essay = EXCLUDED.target_essay,
  status_filter = EXCLUDED.status_filter,
  updated_at = NOW()
RETURNING id, event_id, scope_mode, target_pg, target_essay, status_filter, created_at, updated_at;

-- name: ListCbtEventQuestionCompletenessRows :many
WITH event_scope AS (
  SELECT id, academic_year_id, target_levels
  FROM cbt_exam_events
  WHERE id = $1
), req AS (
  SELECT
    COALESCE(r.scope_mode, 'per_rombel')::text AS scope_mode,
    COALESCE(r.target_pg, 20)::int AS target_pg,
    COALESCE(r.target_essay, 5)::int AS target_essay,
    COALESCE(r.status_filter, 'published_only')::text AS status_filter
  FROM event_scope ev
  LEFT JOIN cbt_event_question_requirements r ON r.event_id = ev.id
    AND r.level IS NULL
    AND r.class_id IS NULL
    AND r.subject_id IS NULL
  ORDER BY r.updated_at DESC NULLS LAST
  LIMIT 1
), target_assignments AS (
  SELECT
    c.level,
    c.id AS class_id,
    c.code AS class_code,
    c.name AS class_name,
    a.subject_id,
    s.name AS subject_name,
    s.code AS subject_code,
    a.teacher_employee_id,
    e.nama AS teacher_name,
    COALESCE(u.username, '')::text AS teacher_username,
    CASE c.level
      WHEN 'VII' THEN 7
      WHEN 'VIII' THEN 8
      WHEN 'IX' THEN 9
      ELSE NULL
    END::smallint AS numeric_grade_level
  FROM event_scope ev
  JOIN school_classes c ON c.academic_year_id = ev.academic_year_id
    AND c.is_active = TRUE
    AND (
      COALESCE(array_length(ev.target_levels, 1), 0) = 0
      OR c.level = ANY(ev.target_levels)
    )
  JOIN class_subject_assignments a ON a.class_id = c.id
  JOIN subjects s ON s.id = a.subject_id AND s.is_active = TRUE
  JOIN employees e ON e.id = a.teacher_employee_id
  LEFT JOIN LATERAL (
    SELECT username
    FROM users ux
    WHERE ux.employee_id = e.id AND ux.deleted_at IS NULL
    ORDER BY ux.created_at DESC
    LIMIT 1
  ) u ON TRUE
), scoped_assignments AS (
  SELECT
    ta.level,
    (CASE WHEN req.scope_mode = 'per_rombel' THEN ta.class_id ELSE NULL::uuid END)::uuid AS class_id,
    (CASE WHEN req.scope_mode = 'per_rombel' THEN ta.class_code ELSE '' END)::text AS class_code,
    (CASE WHEN req.scope_mode = 'per_rombel' THEN ta.class_name ELSE CONCAT('Tingkat ', ta.level) END)::text AS class_name,
    ta.subject_id,
    ta.subject_name,
    ta.subject_code,
    (CASE WHEN req.scope_mode = 'pool_level_subject' THEN NULL::uuid ELSE ta.teacher_employee_id END)::uuid AS teacher_employee_id,
    (CASE WHEN req.scope_mode = 'pool_level_subject' THEN 'Pool guru mapel' ELSE ta.teacher_name END)::text AS teacher_name,
    (CASE WHEN req.scope_mode = 'pool_level_subject' THEN '' ELSE ta.teacher_username END)::text AS teacher_username,
    MAX(ta.numeric_grade_level)::smallint AS numeric_grade_level,
    req.scope_mode,
    req.status_filter,
    req.target_pg,
    req.target_essay
  FROM target_assignments ta
  CROSS JOIN req
  GROUP BY
    ta.level,
    CASE WHEN req.scope_mode = 'per_rombel' THEN ta.class_id ELSE NULL::uuid END,
    CASE WHEN req.scope_mode = 'per_rombel' THEN ta.class_code ELSE '' END,
    CASE WHEN req.scope_mode = 'per_rombel' THEN ta.class_name ELSE CONCAT('Tingkat ', ta.level) END,
    ta.subject_id,
    ta.subject_name,
    ta.subject_code,
    CASE WHEN req.scope_mode = 'pool_level_subject' THEN NULL::uuid ELSE ta.teacher_employee_id END,
    CASE WHEN req.scope_mode = 'pool_level_subject' THEN 'Pool guru mapel' ELSE ta.teacher_name END,
    CASE WHEN req.scope_mode = 'pool_level_subject' THEN '' ELSE ta.teacher_username END,
    req.scope_mode,
    req.status_filter,
    req.target_pg,
    req.target_essay
)
SELECT
  sa.level,
  sa.class_id,
  sa.class_code,
  sa.class_name,
  sa.subject_id,
  sa.subject_name,
  sa.subject_code,
  sa.teacher_employee_id,
  sa.teacher_name,
  sa.teacher_username,
  sa.scope_mode,
  sa.status_filter,
  sa.target_pg,
  COALESCE(COUNT(DISTINCT q.id) FILTER (WHERE q.question_type = 'multiple_choice'), 0)::int AS available_pg,
  sa.target_essay,
  COALESCE(COUNT(DISTINCT q.id) FILTER (WHERE q.question_type = 'essay'), 0)::int AS available_essay
FROM scoped_assignments sa
LEFT JOIN cbt_questions q ON q.subject_id = sa.subject_id
  AND (sa.teacher_username = '' OR q.author_username = sa.teacher_username)
  AND (
    q.event_id = $1
    OR (q.event_id IS NULL AND q.status = 'published')
  )
  AND (sa.numeric_grade_level IS NULL OR q.grade_level = sa.numeric_grade_level)
  AND q.status <> 'archived'
  AND q.workflow_status <> 'rejected'
  AND (sa.status_filter <> 'published_only' OR q.status = 'published')
GROUP BY
  sa.level,
  sa.class_id,
  sa.class_code,
  sa.class_name,
  sa.subject_id,
  sa.subject_name,
  sa.subject_code,
  sa.teacher_employee_id,
  sa.teacher_name,
  sa.teacher_username,
  sa.scope_mode,
  sa.status_filter,
  sa.target_pg,
  sa.target_essay
ORDER BY sa.level ASC, sa.class_name ASC, sa.subject_name ASC, sa.teacher_name ASC;

-- name: ListCbtEventQuestionCompletenessExcludedLevels :many
SELECT DISTINCT c.level
FROM cbt_exam_events e
JOIN school_classes c ON c.academic_year_id = e.academic_year_id
WHERE e.id = $1
  AND c.is_active = TRUE
  AND COALESCE(array_length(e.target_levels, 1), 0) > 0
  AND NOT (c.level = ANY(e.target_levels))
ORDER BY c.level ASC;

-- name: ListCbtEventSessionsReadiness :many
SELECT
  ses.id,
  ses.event_id,
  ses.package_id,
  p.title AS package_title,
  p.subject_id,
  sub.name AS subject_name,
  sub.code AS subject_code,
  ses.class_id,
  COALESCE(c.name, '') AS class_name,
  COALESCE(c.code, '') AS class_code,
  ses.title,
  ses.scheduled_start,
  ses.scheduled_end,
  ses.status,
  COALESCE(pq.question_count, 0)::int AS question_count,
  COALESCE(pq.published_question_count, 0)::int AS published_question_count,
  COALESCE(parts.participant_count, 0)::int AS participant_count,
  COALESCE(parts.assigned_participant_count, 0)::int AS assigned_participant_count,
  COALESCE(parts.missing_seat_count, 0)::int AS missing_seat_count,
  COALESCE(parts.token_ready_count, 0)::int AS token_ready_count,
  COALESCE(parts.joined_count, 0)::int AS joined_count,
  COALESCE(parts.submitted_count, 0)::int AS submitted_count,
  COALESCE(parts.scored_count, 0)::int AS scored_count,
  COALESCE(rooms.room_count, 0)::int AS room_count,
  COALESCE(rooms.total_capacity, 0)::int AS total_capacity,
  COALESCE(rooms.rooms_without_proctor, 0)::int AS rooms_without_proctor,
  COALESCE(rooms.over_capacity_room_count, 0)::int AS over_capacity_room_count,
  COALESCE(rooms.network_not_ready_room_count, 0)::int AS network_not_ready_room_count,
  COALESCE(rooms.power_not_ready_room_count, 0)::int AS power_not_ready_room_count
FROM cbt_exam_sessions ses
JOIN cbt_packages p ON p.id = ses.package_id
JOIN subjects sub ON sub.id = p.subject_id
LEFT JOIN school_classes c ON c.id = ses.class_id
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS question_count, COUNT(*) FILTER (WHERE q.status = 'published')::int AS published_question_count
  FROM cbt_package_questions pq
  JOIN cbt_questions q ON q.id = pq.question_id
  WHERE pq.package_id = p.id
) pq ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COUNT(*)::int AS participant_count,
    COUNT(*) FILTER (WHERE room_id IS NOT NULL)::int AS assigned_participant_count,
    COUNT(*) FILTER (WHERE room_id IS NOT NULL AND seat_no IS NULL)::int AS missing_seat_count,
    COUNT(*) FILTER (WHERE token IS NOT NULL AND token ~ '^[0-9a-f]{32}$')::int AS token_ready_count,
    COUNT(*) FILTER (WHERE joined_at IS NOT NULL)::int AS joined_count,
    COUNT(*) FILTER (WHERE submitted_at IS NOT NULL)::int AS submitted_count,
    COUNT(*) FILTER (WHERE score IS NOT NULL)::int AS scored_count
  FROM cbt_exam_participants ep
  WHERE ep.session_id = ses.id
) parts ON TRUE
LEFT JOIN LATERAL (
  SELECT
    COUNT(r.id)::int AS room_count,
    COALESCE(SUM(COALESCE(r.capacity_override, r.capacity)), 0)::int AS total_capacity,
    COUNT(r.id) FILTER (WHERE COALESCE(pr.proctor_count, 0) = 0)::int AS rooms_without_proctor,
    COUNT(r.id) FILTER (WHERE COALESCE(pc.participant_count, 0) > COALESCE(r.capacity_override, r.capacity))::int AS over_capacity_room_count,
    COUNT(r.id) FILTER (WHERE r.school_room_id IS NOT NULL AND COALESCE(sr.network_ready, FALSE) = FALSE)::int AS network_not_ready_room_count,
    COUNT(r.id) FILTER (WHERE r.school_room_id IS NOT NULL AND COALESCE(sr.power_ready, FALSE) = FALSE)::int AS power_not_ready_room_count
  FROM cbt_exam_rooms r
  LEFT JOIN school_rooms sr ON sr.id = r.school_room_id
  LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS proctor_count
    FROM cbt_room_proctors rp
    WHERE rp.exam_room_id = r.id
  ) pr ON TRUE
  LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS participant_count
    FROM cbt_exam_participants ep
    WHERE ep.session_id = r.session_id AND ep.room_id = r.id
  ) pc ON TRUE
  WHERE r.session_id = ses.id
) rooms ON TRUE
WHERE ses.event_id = $1
ORDER BY ses.scheduled_start ASC, ses.title ASC;

-- name: ListCbtEventMembers :many
SELECT
  m.id, m.event_id, m.user_id, u.username,
  m.employee_id, COALESCE(e.nama, '') AS employee_name, COALESCE(e.nip, '') AS employee_nip,
  m.subject_id, COALESCE(s.name, '') AS subject_name, COALESCE(s.code, '') AS subject_code,
  m.role, m.created_at, m.updated_at
FROM cbt_event_members m
JOIN users u ON u.id = m.user_id
LEFT JOIN employees e ON e.id = m.employee_id
LEFT JOIN subjects s ON s.id = m.subject_id
WHERE m.event_id = $1
ORDER BY m.role ASC, subject_name ASC, u.username ASC;

-- name: GetCbtEventMember :one
SELECT
  m.id, m.event_id, m.user_id, u.username,
  m.employee_id, COALESCE(e.nama, '') AS employee_name, COALESCE(e.nip, '') AS employee_nip,
  m.subject_id, COALESCE(s.name, '') AS subject_name, COALESCE(s.code, '') AS subject_code,
  m.role, m.created_at, m.updated_at
FROM cbt_event_members m
JOIN users u ON u.id = m.user_id
LEFT JOIN employees e ON e.id = m.employee_id
LEFT JOIN subjects s ON s.id = m.subject_id
WHERE m.id = $1 AND m.event_id = $2;

-- name: ListCbtEventMembersByUser :many
SELECT id, event_id, user_id, employee_id, subject_id, role, created_at, updated_at
FROM cbt_event_members
WHERE user_id = $1;

-- name: ListCbtEventMembersByUsername :many
SELECT m.id, m.event_id, m.user_id, m.employee_id, m.subject_id, m.role, m.created_at, m.updated_at
FROM cbt_event_members m
JOIN users u ON u.id = m.user_id
WHERE u.username = $1;

-- name: CreateCbtEventMember :one
INSERT INTO cbt_event_members (event_id, user_id, employee_id, subject_id, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateCbtEventMember :one
UPDATE cbt_event_members
SET user_id = $3,
    employee_id = $4,
    subject_id = $5,
    role = $6,
    updated_at = NOW()
WHERE id = $1 AND event_id = $2
RETURNING *;

-- name: DeleteCbtEventMember :exec
DELETE FROM cbt_event_members WHERE id = $1 AND event_id = $2;

-- name: ListCbtEventSubjectTargets :many
SELECT
  t.id,
  t.event_id,
  t.subject_id,
  s.name AS subject_name,
  s.code AS subject_code,
  t.target_questions,
  COALESCE(progress.total_count, 0)::int AS total_count,
  COALESCE(progress.draft_count, 0)::int AS draft_count,
  COALESCE(progress.review_count, 0)::int AS review_count,
  COALESCE(progress.rejected_count, 0)::int AS rejected_count,
  COALESCE(progress.approved_count, 0)::int AS approved_count,
  COALESCE(progress.published_count, 0)::int AS published_count,
  t.created_at,
  t.updated_at
FROM cbt_event_subject_targets t
JOIN subjects s ON s.id = t.subject_id
LEFT JOIN LATERAL (
  SELECT
    COUNT(*)::int AS total_count,
    COUNT(*) FILTER (WHERE q.workflow_status = 'draft')::int AS draft_count,
    COUNT(*) FILTER (WHERE q.workflow_status = 'review')::int AS review_count,
    COUNT(*) FILTER (WHERE q.workflow_status = 'rejected')::int AS rejected_count,
    COUNT(*) FILTER (WHERE q.workflow_status = 'approved')::int AS approved_count,
    COUNT(*) FILTER (WHERE q.status = 'published')::int AS published_count
  FROM cbt_questions q
  WHERE q.subject_id = t.subject_id
    AND (
      q.event_id = t.event_id
      OR (q.event_id IS NULL AND q.status = 'published')
    )
) progress ON TRUE
WHERE t.event_id = $1
ORDER BY s.name ASC, s.code ASC;

-- name: UpsertCbtEventSubjectTarget :one
INSERT INTO cbt_event_subject_targets (event_id, subject_id, target_questions)
VALUES ($1, $2, $3)
ON CONFLICT (event_id, subject_id)
DO UPDATE SET target_questions = EXCLUDED.target_questions, updated_at = NOW()
RETURNING *;

-- name: DeleteCbtEventSubjectTarget :execrows
DELETE FROM cbt_event_subject_targets
WHERE event_id = $1 AND subject_id = $2;

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
