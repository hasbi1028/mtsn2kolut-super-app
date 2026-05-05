-- name: ListCbtQuestions :many
SELECT q.id, q.event_id, q.subject_id, s.name AS subject_name, s.code AS subject_code,
       q.code, q.question_text, q.question_type, q.options,
       q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
       CASE
         WHEN sqlc.arg(is_admin)::bool
           OR q.author_username = sqlc.arg(actor_username)::text
           OR EXISTS (
             SELECT 1 FROM cbt_event_members m
             WHERE m.event_id = q.event_id
               AND m.user_id = sqlc.arg(actor_user_id)::uuid
               AND m.role IN ('reviewer', 'panitia')
               AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
           )
         THEN q.answer_key ELSE '' END AS answer_key,
       q.explanation, q.difficulty, q.status, q.created_at, q.updated_at,
       q.stem_html, q.stem_latex, q.stimulus_html, q.stimulus_latex,
       q.explanation_html, q.rubric_html,
       q.academic_phase, q.grade_level,
       q.cp_ref, q.tp_ref, q.kd_ref, q.indicator_ref,
       q.material_topic, q.cognitive_level, q.hots_flag,
       q.media_asset_ids, q.workflow_status, q.version,
       q.author_username, q.reviewer_username, q.reviewed_at,
       q.approver_username, q.approved_at, q.writer_notes, q.review_notes,
       COALESCE(pkg_usage.package_count, 0)::int AS package_count,
       COALESCE(answer_usage.answer_count, 0)::int AS answer_count
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS package_count
  FROM cbt_package_questions pq
  WHERE pq.question_id = q.id
) pkg_usage ON TRUE
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS answer_count
  FROM cbt_student_answers sa
  WHERE sa.question_id = q.id
) answer_usage ON TRUE
ORDER BY q.created_at DESC;

-- name: ListCbtQuestionsFiltered :many
SELECT q.id, q.event_id, q.subject_id, s.name AS subject_name, s.code AS subject_code,
       q.code, q.question_text, q.question_type, q.options,
       q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
       CASE
         WHEN sqlc.arg(is_admin)::bool
           OR q.author_username = sqlc.arg(actor_username)::text
           OR EXISTS (
             SELECT 1 FROM cbt_event_members m
             WHERE m.event_id = q.event_id
               AND m.user_id = sqlc.arg(actor_user_id)::uuid
               AND m.role IN ('reviewer', 'panitia')
               AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
           )
         THEN q.answer_key ELSE '' END AS answer_key,
       q.explanation, q.difficulty, q.status, q.created_at, q.updated_at,
       q.stem_html, q.stem_latex, q.stimulus_html, q.stimulus_latex,
       q.explanation_html, q.rubric_html,
       q.academic_phase, q.grade_level,
       q.cp_ref, q.tp_ref, q.kd_ref, q.indicator_ref,
       q.material_topic, q.cognitive_level, q.hots_flag,
       q.media_asset_ids, q.workflow_status, q.version,
       q.author_username, q.reviewer_username, q.reviewed_at,
       q.approver_username, q.approved_at, q.writer_notes, q.review_notes,
       COALESCE(pkg_usage.package_count, 0)::int AS package_count,
       COALESCE(answer_usage.answer_count, 0)::int AS answer_count
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS package_count
  FROM cbt_package_questions pq
  WHERE pq.question_id = q.id
) pkg_usage ON TRUE
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS answer_count
  FROM cbt_student_answers sa
  WHERE sa.question_id = q.id
) answer_usage ON TRUE
WHERE (
    (sqlc.arg(scope_filter)::text = 'global' AND q.event_id IS NULL)
    OR (
      sqlc.arg(scope_filter)::text = 'event_pool'
      AND (q.event_id IS NULL OR (sqlc.arg(event_id)::uuid IS NOT NULL AND q.event_id = sqlc.arg(event_id)::uuid))
    )
    OR (
      sqlc.arg(scope_filter)::text NOT IN ('global', 'event_pool')
      AND (sqlc.arg(event_id)::uuid IS NULL OR q.event_id = sqlc.arg(event_id)::uuid)
    )
  )
  AND (sqlc.arg(subject_id)::uuid IS NULL OR q.subject_id = sqlc.arg(subject_id)::uuid)
  AND (sqlc.arg(author_username)::text = '' OR q.author_username = sqlc.arg(author_username)::text)
  AND (sqlc.arg(workflow_status)::text = '' OR q.workflow_status = sqlc.arg(workflow_status)::text)
  AND (sqlc.arg(status_filter)::text = '' OR q.status = sqlc.arg(status_filter)::cbt_question_status_enum)
  AND (sqlc.arg(question_type)::text = '' OR q.question_type = sqlc.arg(question_type)::text)
  AND (sqlc.arg(hots_filter)::text = '' OR (sqlc.arg(hots_filter)::text = 'yes' AND q.hots_flag = TRUE) OR (sqlc.arg(hots_filter)::text = 'no' AND q.hots_flag = FALSE))
  AND (
    sqlc.arg(is_admin)::bool
    OR q.status = 'published'
    OR q.author_username = sqlc.arg(actor_username)::text
    OR EXISTS (
      SELECT 1 FROM cbt_event_members m
      WHERE m.event_id = q.event_id
        AND m.user_id = sqlc.arg(actor_user_id)::uuid
        AND m.role IN ('reviewer', 'panitia')
        AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
    )
  )
  AND (
    sqlc.arg(revision_source)::text = ''
    OR (
      sqlc.arg(revision_source)::text = 'item_analysis'
      AND q.workflow_status = 'rejected'
      AND q.review_notes ILIKE '%analisis butir%'
    )
    OR (
      sqlc.arg(revision_source)::text = 'reviewer'
      AND q.workflow_status = 'rejected'
      AND q.review_notes NOT ILIKE '%analisis butir%'
      AND btrim(q.reviewer_username) <> ''
    )
    OR (
      sqlc.arg(revision_source)::text = 'workflow'
      AND q.workflow_status = 'rejected'
      AND q.review_notes NOT ILIKE '%analisis butir%'
      AND btrim(q.reviewer_username) = ''
    )
  )
  AND (
    sqlc.arg(search_query)::text = ''
    OR q.code ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.question_text ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.material_topic ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.cp_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.kd_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
  )
ORDER BY q.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: ListCbtQuestionsScoped :many
SELECT q.id, q.event_id, q.subject_id, s.name AS subject_name, s.code AS subject_code,
       q.code, q.question_text, q.question_type, q.options,
       q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
       CASE
         WHEN sqlc.arg(is_admin)::bool
           OR q.author_username = sqlc.arg(actor_username)::text
           OR EXISTS (
             SELECT 1 FROM cbt_event_members m
             WHERE m.event_id = q.event_id
               AND m.user_id = sqlc.arg(actor_user_id)::uuid
               AND m.role IN ('reviewer', 'panitia')
               AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
           )
         THEN q.answer_key ELSE '' END AS answer_key,
       q.explanation, q.difficulty, q.status, q.created_at, q.updated_at,
       q.stem_html, q.stem_latex, q.stimulus_html, q.stimulus_latex,
       q.explanation_html, q.rubric_html,
       q.academic_phase, q.grade_level,
       q.cp_ref, q.tp_ref, q.kd_ref, q.indicator_ref,
       q.material_topic, q.cognitive_level, q.hots_flag,
       q.media_asset_ids, q.workflow_status, q.version,
       q.author_username, q.reviewer_username, q.reviewed_at,
       q.approver_username, q.approved_at, q.writer_notes, q.review_notes,
       COALESCE(pkg_usage.package_count, 0)::int AS package_count,
       COALESCE(answer_usage.answer_count, 0)::int AS answer_count
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS package_count
  FROM cbt_package_questions pq
  WHERE pq.question_id = q.id
) pkg_usage ON TRUE
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS answer_count
  FROM cbt_student_answers sa
  WHERE sa.question_id = q.id
) answer_usage ON TRUE
WHERE (
    (sqlc.arg(scope_filter)::text = 'global' AND q.event_id IS NULL)
    OR (
      sqlc.arg(scope_filter)::text = 'event_pool'
      AND (q.event_id IS NULL OR (sqlc.arg(event_id)::uuid IS NOT NULL AND q.event_id = sqlc.arg(event_id)::uuid))
    )
    OR (
      sqlc.arg(scope_filter)::text NOT IN ('global', 'event_pool')
      AND (sqlc.arg(event_id)::uuid IS NULL OR q.event_id = sqlc.arg(event_id)::uuid)
    )
  )
  AND (sqlc.arg(subject_id)::uuid IS NULL OR q.subject_id = sqlc.arg(subject_id)::uuid)
  AND (sqlc.arg(author_username)::text = '' OR q.author_username = sqlc.arg(author_username)::text)
  AND (sqlc.arg(workflow_status)::text = '' OR q.workflow_status = sqlc.arg(workflow_status)::text)
  AND (sqlc.arg(status_filter)::text = '' OR q.status = sqlc.arg(status_filter)::cbt_question_status_enum)
  AND (sqlc.arg(question_type)::text = '' OR q.question_type = sqlc.arg(question_type)::text)
  AND (sqlc.arg(hots_filter)::text = '' OR (sqlc.arg(hots_filter)::text = 'yes' AND q.hots_flag = TRUE) OR (sqlc.arg(hots_filter)::text = 'no' AND q.hots_flag = FALSE))
  AND (
    sqlc.arg(is_admin)::bool
    OR q.status = 'published'
    OR q.author_username = sqlc.arg(actor_username)::text
    OR EXISTS (
      SELECT 1 FROM cbt_event_members m
      WHERE m.event_id = q.event_id
        AND m.user_id = sqlc.arg(actor_user_id)::uuid
        AND m.role IN ('reviewer', 'panitia')
        AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
    )
  )
  AND (
    sqlc.arg(revision_source)::text = ''
    OR (
      sqlc.arg(revision_source)::text = 'item_analysis'
      AND q.workflow_status = 'rejected'
      AND q.review_notes ILIKE '%analisis butir%'
    )
    OR (
      sqlc.arg(revision_source)::text = 'reviewer'
      AND q.workflow_status = 'rejected'
      AND q.review_notes NOT ILIKE '%analisis butir%'
      AND btrim(q.reviewer_username) <> ''
    )
    OR (
      sqlc.arg(revision_source)::text = 'workflow'
      AND q.workflow_status = 'rejected'
      AND q.review_notes NOT ILIKE '%analisis butir%'
      AND btrim(q.reviewer_username) = ''
    )
  )
  AND (
    sqlc.arg(search_query)::text = ''
    OR q.code ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.question_text ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.material_topic ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.cp_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.kd_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
  )
ORDER BY q.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountCbtQuestionsFiltered :one
SELECT COUNT(*)::bigint
FROM cbt_questions q
WHERE (
    (sqlc.arg(scope_filter)::text = 'global' AND q.event_id IS NULL)
    OR (
      sqlc.arg(scope_filter)::text = 'event_pool'
      AND (q.event_id IS NULL OR (sqlc.arg(event_id)::uuid IS NOT NULL AND q.event_id = sqlc.arg(event_id)::uuid))
    )
    OR (
      sqlc.arg(scope_filter)::text NOT IN ('global', 'event_pool')
      AND (sqlc.arg(event_id)::uuid IS NULL OR q.event_id = sqlc.arg(event_id)::uuid)
    )
  )
  AND (sqlc.arg(subject_id)::uuid IS NULL OR q.subject_id = sqlc.arg(subject_id)::uuid)
  AND (sqlc.arg(author_username)::text = '' OR q.author_username = sqlc.arg(author_username)::text)
  AND (sqlc.arg(workflow_status)::text = '' OR q.workflow_status = sqlc.arg(workflow_status)::text)
  AND (sqlc.arg(status_filter)::text = '' OR q.status = sqlc.arg(status_filter)::cbt_question_status_enum)
  AND (sqlc.arg(question_type)::text = '' OR q.question_type = sqlc.arg(question_type)::text)
  AND (sqlc.arg(hots_filter)::text = '' OR (sqlc.arg(hots_filter)::text = 'yes' AND q.hots_flag = TRUE) OR (sqlc.arg(hots_filter)::text = 'no' AND q.hots_flag = FALSE))
  AND (
    sqlc.arg(is_admin)::bool
    OR q.status = 'published'
    OR q.author_username = sqlc.arg(actor_username)::text
    OR EXISTS (
      SELECT 1 FROM cbt_event_members m
      WHERE m.event_id = q.event_id
        AND m.user_id = sqlc.arg(actor_user_id)::uuid
        AND m.role IN ('reviewer', 'panitia')
        AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
    )
  )
  AND (
    sqlc.arg(revision_source)::text = ''
    OR (
      sqlc.arg(revision_source)::text = 'item_analysis'
      AND q.workflow_status = 'rejected'
      AND q.review_notes ILIKE '%analisis butir%'
    )
    OR (
      sqlc.arg(revision_source)::text = 'reviewer'
      AND q.workflow_status = 'rejected'
      AND q.review_notes NOT ILIKE '%analisis butir%'
      AND btrim(q.reviewer_username) <> ''
    )
    OR (
      sqlc.arg(revision_source)::text = 'workflow'
      AND q.workflow_status = 'rejected'
      AND q.review_notes NOT ILIKE '%analisis butir%'
      AND btrim(q.reviewer_username) = ''
    )
  )
  AND (
    sqlc.arg(search_query)::text = ''
    OR q.code ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.question_text ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.material_topic ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.cp_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.kd_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
  );

-- name: GetCbtQuestion :one
SELECT q.id, q.event_id, q.subject_id, q.code, q.question_text, q.question_type, q.options,
       q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
       q.answer_key, q.explanation, q.difficulty, q.status, q.created_at, q.updated_at,
       q.stem_html, q.stem_latex, q.stimulus_html, q.stimulus_latex,
       q.explanation_html, q.rubric_html,
       q.academic_phase, q.grade_level,
       q.cp_ref, q.tp_ref, q.kd_ref, q.indicator_ref,
       q.material_topic, q.cognitive_level, q.hots_flag,
       q.media_asset_ids, q.workflow_status, q.version,
       q.author_username, q.reviewer_username, q.reviewed_at,
       q.approver_username, q.approved_at, q.writer_notes, q.review_notes,
       COALESCE(pkg_usage.package_count, 0)::int AS package_count,
       COALESCE(answer_usage.answer_count, 0)::int AS answer_count
FROM cbt_questions q
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS package_count
  FROM cbt_package_questions pq
  WHERE pq.question_id = q.id
) pkg_usage ON TRUE
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS answer_count
  FROM cbt_student_answers sa
  WHERE sa.question_id = q.id
) answer_usage ON TRUE
WHERE q.id = $1;

-- name: GetCbtQuestionDetail :one
SELECT q.id, q.event_id, q.subject_id, s.name AS subject_name, s.code AS subject_code,
       q.code, q.question_text, q.question_type, q.options,
       q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
       q.answer_key, q.explanation, q.difficulty, q.status, q.created_at, q.updated_at,
       q.stem_html, q.stem_latex, q.stimulus_html, q.stimulus_latex,
       q.explanation_html, q.rubric_html,
       q.academic_phase, q.grade_level,
       q.cp_ref, q.tp_ref, q.kd_ref, q.indicator_ref,
       q.material_topic, q.cognitive_level, q.hots_flag,
       q.media_asset_ids, q.workflow_status, q.version,
       q.author_username, q.reviewer_username, q.reviewed_at,
       q.approver_username, q.approved_at, q.writer_notes, q.review_notes,
       COALESCE(pkg_usage.package_count, 0)::int AS package_count,
       COALESCE(answer_usage.answer_count, 0)::int AS answer_count
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS package_count
  FROM cbt_package_questions pq
  WHERE pq.question_id = q.id
) pkg_usage ON TRUE
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS answer_count
  FROM cbt_student_answers sa
  WHERE sa.question_id = q.id
) answer_usage ON TRUE
WHERE q.id = $1;

-- name: ListCbtQuestionStemTextsBySubject :many
SELECT question_text, stem_html
FROM cbt_questions
WHERE subject_id = $1;

-- name: CreateCbtQuestion :one
INSERT INTO cbt_questions (
  id, event_id, subject_id, code, question_text, question_type, options,
  option_a, option_b, option_c, option_d, option_e,
  answer_key, explanation, difficulty, status,
  stem_html, stem_latex, stimulus_html, stimulus_latex,
  explanation_html, rubric_html,
  academic_phase, grade_level,
  cp_ref, tp_ref, kd_ref, indicator_ref,
  material_topic, cognitive_level, hots_flag,
  media_asset_ids, workflow_status, version,
  author_username, reviewer_username, reviewed_at,
  approver_username, approved_at, writer_notes, review_notes
)
VALUES (
  gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
  $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31,
  $32, $33, $34, $35, $36, $37, $38, $39, $40
)
RETURNING *;

-- name: UpdateCbtQuestion :one
UPDATE cbt_questions
SET
  event_id           = $2,
  subject_id         = $3,
  code               = $4,
  question_text      = $5,
  question_type      = $6,
  options            = $7,
  option_a           = $8,
  option_b           = $9,
  option_c           = $10,
  option_d           = $11,
  option_e           = $12,
  answer_key         = $13,
  explanation        = $14,
  difficulty         = $15,
  status             = $16,
  stem_html          = $17,
  stem_latex         = $18,
  stimulus_html      = $19,
  stimulus_latex     = $20,
  explanation_html   = $21,
  rubric_html        = $22,
  academic_phase     = $23,
  grade_level        = $24,
  cp_ref             = $25,
  tp_ref             = $26,
  kd_ref             = $27,
  indicator_ref      = $28,
  material_topic     = $29,
  cognitive_level    = $30,
  hots_flag          = $31,
  media_asset_ids    = $32,
  workflow_status    = $33,
  version            = version + 1,
  reviewer_username  = $34,
  reviewed_at        = $35,
  approver_username  = $36,
  approved_at        = $37,
  writer_notes       = $38,
  review_notes       = $39,
  updated_at         = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCbtQuestion :exec
DELETE FROM cbt_questions WHERE id = $1;

-- name: CreateCbtQuestionAuditLog :one
INSERT INTO cbt_question_audit_logs (question_id, actor_username, action, note, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListCbtQuestionTimeline :many
SELECT id, question_id, actor_username, action, note, metadata, created_at
FROM cbt_question_audit_logs
WHERE question_id = $1
ORDER BY created_at ASC, id ASC;

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
  q.stem_html,
  q.stimulus_html,
  q.rubric_html,
  COALESCE(pq.points, 1)::numeric AS points,
  s.nis, s.nama,
  COALESCE(r.room_name, '') AS room_name
FROM cbt_student_answers sa
JOIN cbt_questions q ON q.id = sa.question_id
JOIN cbt_exam_participants ep ON ep.id = sa.participant_id
JOIN cbt_exam_sessions ses ON ses.id = ep.session_id
LEFT JOIN cbt_package_questions pq ON pq.package_id = ses.package_id AND pq.question_id = q.id
JOIN students s ON s.id = ep.student_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.session_id = $1
  AND q.question_type = 'essay'
  AND sa.manual_score IS NULL
ORDER BY q.code, s.nama;

-- name: ListUngradedEssaysByTeacher :many
SELECT
  sa.id AS answer_id,
  sa.participant_id,
  sa.question_id,
  sa.answer,
  sa.manual_score,
  sa.graded_at,
  q.code AS question_code,
  q.question_text,
  q.stem_html,
  q.stimulus_html,
  q.rubric_html,
  COALESCE(pq.points, 1)::numeric AS points,
  s.nis, s.nama,
  COALESCE(r.room_name, '') AS room_name
FROM cbt_student_answers sa
JOIN cbt_questions q ON q.id = sa.question_id
JOIN cbt_exam_participants ep ON ep.id = sa.participant_id
JOIN cbt_exam_sessions ses ON ses.id = ep.session_id
JOIN cbt_packages pkg ON pkg.id = ses.package_id
LEFT JOIN cbt_package_questions pq ON pq.package_id = ses.package_id AND pq.question_id = q.id
JOIN students s ON s.id = ep.student_id
JOIN class_subject_assignments csa
  ON csa.subject_id = pkg.subject_id
 AND csa.class_id = s.class_id
 AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.session_id = sqlc.arg(session_id)
  AND q.question_type = 'essay'
  AND sa.manual_score IS NULL
ORDER BY q.code, s.nama;
