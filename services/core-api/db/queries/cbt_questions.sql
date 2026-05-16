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
           OR EXISTS (
             SELECT 1 FROM bank_soal_reviewer_scopes rs
             WHERE rs.user_id = sqlc.arg(actor_user_id)::uuid
               AND (
                 (rs.can_review = TRUE AND q.workflow_status IN ('submitted', 'review', 'revision_needed', 'reviewed'))
                 OR (rs.can_approve = TRUE AND q.workflow_status IN ('reviewed', 'approved', 'published'))
               )
               AND (rs.subject_id IS NULL OR rs.subject_id = q.subject_id)
               AND (
                 rs.grade_level IS NULL
                 OR rs.grade_level = CASE q.target_level WHEN 'VII' THEN 7 WHEN 'VIII' THEN 8 WHEN 'IX' THEN 9 ELSE NULL END
               )
           )
         THEN q.answer_key ELSE '' END AS answer_key,
       q.explanation, q.difficulty, q.status, q.created_at, q.updated_at,
       q.stem_html, q.stem_latex, q.stimulus_html, q.stimulus_latex,
       q.explanation_html, q.rubric_html,
       q.academic_phase, q.target_level,
       q.cp_ref, q.tp_ref, q.kd_ref, q.indicator_ref,
       q.material_topic, q.cognitive_level, q.hots_flag,
       q.media_asset_ids, q.workflow_status, q.version,
       q.version_group_id, q.version_number, q.source_question_id,
       q.supersedes_question_id, q.is_latest_version, q.version_note,
       q.author_username,
       COALESCE(NULLIF(btrim(author_emp.nama), ''), NULLIF(btrim(author_user.display_name), ''), q.author_username) AS author_display_name,
       q.reviewer_username,
       COALESCE(NULLIF(btrim(reviewer_emp.nama), ''), NULLIF(btrim(reviewer_user.display_name), ''), q.reviewer_username) AS reviewer_display_name,
       q.reviewed_at,
       q.approver_username,
       COALESCE(NULLIF(btrim(approver_emp.nama), ''), NULLIF(btrim(approver_user.display_name), ''), q.approver_username) AS approver_display_name,
       q.approved_at, q.writer_notes, q.review_notes,
       COALESCE(pkg_usage.package_count, 0)::int AS package_count,
       COALESCE(answer_usage.answer_count, 0)::int AS answer_count
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
LEFT JOIN users author_user ON author_user.username = q.author_username
LEFT JOIN employees author_emp ON author_emp.id = author_user.employee_id
LEFT JOIN users reviewer_user ON reviewer_user.username = q.reviewer_username
LEFT JOIN employees reviewer_emp ON reviewer_emp.id = reviewer_user.employee_id
LEFT JOIN users approver_user ON approver_user.username = q.approver_username
LEFT JOIN employees approver_emp ON approver_emp.id = approver_user.employee_id
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
       q.academic_phase, q.target_level,
       q.cp_ref, q.tp_ref, q.kd_ref, q.indicator_ref,
       q.material_topic, q.cognitive_level, q.hots_flag,
       q.media_asset_ids, q.workflow_status, q.version,
       q.version_group_id, q.version_number, q.source_question_id,
       q.supersedes_question_id, q.is_latest_version, q.version_note,
       q.author_username,
       COALESCE(NULLIF(btrim(author_emp.nama), ''), NULLIF(btrim(author_user.display_name), ''), q.author_username) AS author_display_name,
       q.reviewer_username,
       COALESCE(NULLIF(btrim(reviewer_emp.nama), ''), NULLIF(btrim(reviewer_user.display_name), ''), q.reviewer_username) AS reviewer_display_name,
       q.reviewed_at,
       q.approver_username,
       COALESCE(NULLIF(btrim(approver_emp.nama), ''), NULLIF(btrim(approver_user.display_name), ''), q.approver_username) AS approver_display_name,
       q.approved_at, q.writer_notes, q.review_notes,
       COALESCE(pkg_usage.package_count, 0)::int AS package_count,
       COALESCE(answer_usage.answer_count, 0)::int AS answer_count
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
LEFT JOIN users author_user ON author_user.username = q.author_username
LEFT JOIN employees author_emp ON author_emp.id = author_user.employee_id
LEFT JOIN users reviewer_user ON reviewer_user.username = q.reviewer_username
LEFT JOIN employees reviewer_emp ON reviewer_emp.id = reviewer_user.employee_id
LEFT JOIN users approver_user ON approver_user.username = q.approver_username
LEFT JOIN employees approver_emp ON approver_emp.id = approver_user.employee_id
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
  AND (
    sqlc.arg(workflow_status)::text = ''
    OR q.workflow_status = sqlc.arg(workflow_status)::text
    OR (sqlc.arg(workflow_status)::text = 'submitted' AND q.workflow_status = 'review')
    OR (sqlc.arg(workflow_status)::text = 'review' AND q.workflow_status = 'submitted')
  )
  AND (sqlc.arg(status_filter)::text = '' OR q.status = sqlc.arg(status_filter)::cbt_question_status_enum)
  AND (sqlc.arg(question_type)::text = '' OR q.question_type = sqlc.arg(question_type)::text)
  AND (sqlc.arg(target_level)::text = '' OR q.target_level = sqlc.arg(target_level)::text)
  AND (sqlc.arg(difficulty_filter)::text = '' OR q.difficulty = sqlc.arg(difficulty_filter)::cbt_question_difficulty_enum)
  AND (sqlc.arg(cognitive_level)::text = '' OR q.cognitive_level = sqlc.arg(cognitive_level)::text)
  AND (sqlc.arg(material_topic)::text = '' OR q.material_topic ILIKE '%' || sqlc.arg(material_topic)::text || '%')
  AND (
    sqlc.arg(metadata_filter)::text = ''
    OR (
      sqlc.arg(metadata_filter)::text = 'complete'
      AND NULLIF(btrim(COALESCE(q.target_level, '')), '') IS NOT NULL
      AND NULLIF(btrim(q.cp_ref), '') IS NOT NULL
      AND (NULLIF(btrim(q.tp_ref), '') IS NOT NULL OR NULLIF(btrim(q.kd_ref), '') IS NOT NULL)
      AND NULLIF(btrim(q.cognitive_level), '') IS NOT NULL
    )
    OR (
      sqlc.arg(metadata_filter)::text = 'gap'
      AND (
        NULLIF(btrim(COALESCE(q.target_level, '')), '') IS NULL
        OR NULLIF(btrim(q.cp_ref), '') IS NULL
        OR (NULLIF(btrim(q.tp_ref), '') IS NULL AND NULLIF(btrim(q.kd_ref), '') IS NULL)
        OR NULLIF(btrim(q.cognitive_level), '') IS NULL
      )
    )
  )
  AND (sqlc.arg(hots_filter)::text = '' OR (sqlc.arg(hots_filter)::text = 'yes' AND q.hots_flag = TRUE) OR (sqlc.arg(hots_filter)::text = 'no' AND q.hots_flag = FALSE))
  AND (
    sqlc.arg(is_admin)::bool
    OR q.status = 'published'
    OR (sqlc.arg(can_use_in_package)::bool AND q.workflow_status IN ('approved', 'published'))
    OR q.author_username = sqlc.arg(actor_username)::text
    OR EXISTS (
      SELECT 1 FROM cbt_event_members m
      WHERE m.event_id = q.event_id
        AND m.user_id = sqlc.arg(actor_user_id)::uuid
        AND m.role IN ('reviewer', 'panitia')
        AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
    )
    OR EXISTS (
      SELECT 1 FROM bank_soal_reviewer_scopes rs
      WHERE rs.user_id = sqlc.arg(actor_user_id)::uuid
        AND (
          (rs.can_review = TRUE AND q.workflow_status IN ('submitted', 'review', 'revision_needed', 'reviewed'))
          OR (rs.can_approve = TRUE AND q.workflow_status IN ('reviewed', 'approved', 'published'))
        )
        AND (rs.subject_id IS NULL OR rs.subject_id = q.subject_id)
        AND (
          rs.grade_level IS NULL
          OR rs.grade_level = CASE q.target_level WHEN 'VII' THEN 7 WHEN 'VIII' THEN 8 WHEN 'IX' THEN 9 ELSE NULL END
        )
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
    OR q.tp_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.kd_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.indicator_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
  )
ORDER BY
  CASE WHEN sqlc.arg(sort_order)::text = 'code_asc' THEN q.code END ASC,
  CASE WHEN sqlc.arg(sort_order)::text = 'updated_desc' THEN q.updated_at END DESC,
  CASE WHEN sqlc.arg(sort_order)::text = 'created_asc' THEN q.created_at END ASC,
  CASE WHEN sqlc.arg(sort_order)::text = 'difficulty_asc' THEN q.difficulty::text END ASC,
  CASE WHEN sqlc.arg(sort_order)::text = 'type_asc' THEN q.question_type END ASC,
  q.created_at DESC
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
       q.academic_phase, q.target_level,
       q.cp_ref, q.tp_ref, q.kd_ref, q.indicator_ref,
       q.material_topic, q.cognitive_level, q.hots_flag,
       q.media_asset_ids, q.workflow_status, q.version,
       q.version_group_id, q.version_number, q.source_question_id,
       q.supersedes_question_id, q.is_latest_version, q.version_note,
       q.author_username,
       COALESCE(NULLIF(btrim(author_emp.nama), ''), NULLIF(btrim(author_user.display_name), ''), q.author_username) AS author_display_name,
       q.reviewer_username,
       COALESCE(NULLIF(btrim(reviewer_emp.nama), ''), NULLIF(btrim(reviewer_user.display_name), ''), q.reviewer_username) AS reviewer_display_name,
       q.reviewed_at,
       q.approver_username,
       COALESCE(NULLIF(btrim(approver_emp.nama), ''), NULLIF(btrim(approver_user.display_name), ''), q.approver_username) AS approver_display_name,
       q.approved_at, q.writer_notes, q.review_notes,
       COALESCE(pkg_usage.package_count, 0)::int AS package_count,
       COALESCE(answer_usage.answer_count, 0)::int AS answer_count
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
LEFT JOIN users author_user ON author_user.username = q.author_username
LEFT JOIN employees author_emp ON author_emp.id = author_user.employee_id
LEFT JOIN users reviewer_user ON reviewer_user.username = q.reviewer_username
LEFT JOIN employees reviewer_emp ON reviewer_emp.id = reviewer_user.employee_id
LEFT JOIN users approver_user ON approver_user.username = q.approver_username
LEFT JOIN employees approver_emp ON approver_emp.id = approver_user.employee_id
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
  AND (
    sqlc.arg(workflow_status)::text = ''
    OR q.workflow_status = sqlc.arg(workflow_status)::text
    OR (sqlc.arg(workflow_status)::text = 'submitted' AND q.workflow_status = 'review')
    OR (sqlc.arg(workflow_status)::text = 'review' AND q.workflow_status = 'submitted')
  )
  AND (sqlc.arg(status_filter)::text = '' OR q.status = sqlc.arg(status_filter)::cbt_question_status_enum)
  AND (sqlc.arg(question_type)::text = '' OR q.question_type = sqlc.arg(question_type)::text)
  AND (sqlc.arg(target_level)::text = '' OR q.target_level = sqlc.arg(target_level)::text)
  AND (sqlc.arg(difficulty_filter)::text = '' OR q.difficulty = sqlc.arg(difficulty_filter)::cbt_question_difficulty_enum)
  AND (sqlc.arg(cognitive_level)::text = '' OR q.cognitive_level = sqlc.arg(cognitive_level)::text)
  AND (sqlc.arg(material_topic)::text = '' OR q.material_topic ILIKE '%' || sqlc.arg(material_topic)::text || '%')
  AND (
    sqlc.arg(metadata_filter)::text = ''
    OR (
      sqlc.arg(metadata_filter)::text = 'complete'
      AND NULLIF(btrim(COALESCE(q.target_level, '')), '') IS NOT NULL
      AND NULLIF(btrim(q.cp_ref), '') IS NOT NULL
      AND (NULLIF(btrim(q.tp_ref), '') IS NOT NULL OR NULLIF(btrim(q.kd_ref), '') IS NOT NULL)
      AND NULLIF(btrim(q.cognitive_level), '') IS NOT NULL
    )
    OR (
      sqlc.arg(metadata_filter)::text = 'gap'
      AND (
        NULLIF(btrim(COALESCE(q.target_level, '')), '') IS NULL
        OR NULLIF(btrim(q.cp_ref), '') IS NULL
        OR (NULLIF(btrim(q.tp_ref), '') IS NULL AND NULLIF(btrim(q.kd_ref), '') IS NULL)
        OR NULLIF(btrim(q.cognitive_level), '') IS NULL
      )
    )
  )
  AND (sqlc.arg(hots_filter)::text = '' OR (sqlc.arg(hots_filter)::text = 'yes' AND q.hots_flag = TRUE) OR (sqlc.arg(hots_filter)::text = 'no' AND q.hots_flag = FALSE))
  AND (
    sqlc.arg(is_admin)::bool
    OR q.status = 'published'
    OR (sqlc.arg(can_use_in_package)::bool AND q.workflow_status IN ('approved', 'published'))
    OR q.author_username = sqlc.arg(actor_username)::text
    OR EXISTS (
      SELECT 1 FROM cbt_event_members m
      WHERE m.event_id = q.event_id
        AND m.user_id = sqlc.arg(actor_user_id)::uuid
        AND m.role IN ('reviewer', 'panitia')
        AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
    )
    OR EXISTS (
      SELECT 1 FROM bank_soal_reviewer_scopes rs
      WHERE rs.user_id = sqlc.arg(actor_user_id)::uuid
        AND (
          (rs.can_review = TRUE AND q.workflow_status IN ('submitted', 'review', 'revision_needed', 'reviewed'))
          OR (rs.can_approve = TRUE AND q.workflow_status IN ('reviewed', 'approved', 'published'))
        )
        AND (rs.subject_id IS NULL OR rs.subject_id = q.subject_id)
        AND (
          rs.grade_level IS NULL
          OR rs.grade_level = CASE q.target_level WHEN 'VII' THEN 7 WHEN 'VIII' THEN 8 WHEN 'IX' THEN 9 ELSE NULL END
        )
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
    OR q.tp_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.kd_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.indicator_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
  )
ORDER BY
  CASE WHEN sqlc.arg(sort_order)::text = 'code_asc' THEN q.code END ASC,
  CASE WHEN sqlc.arg(sort_order)::text = 'updated_desc' THEN q.updated_at END DESC,
  CASE WHEN sqlc.arg(sort_order)::text = 'created_asc' THEN q.created_at END ASC,
  CASE WHEN sqlc.arg(sort_order)::text = 'difficulty_asc' THEN q.difficulty::text END ASC,
  CASE WHEN sqlc.arg(sort_order)::text = 'type_asc' THEN q.question_type END ASC,
  q.created_at DESC
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
  AND (
    sqlc.arg(workflow_status)::text = ''
    OR q.workflow_status = sqlc.arg(workflow_status)::text
    OR (sqlc.arg(workflow_status)::text = 'submitted' AND q.workflow_status = 'review')
    OR (sqlc.arg(workflow_status)::text = 'review' AND q.workflow_status = 'submitted')
  )
  AND (sqlc.arg(status_filter)::text = '' OR q.status = sqlc.arg(status_filter)::cbt_question_status_enum)
  AND (sqlc.arg(question_type)::text = '' OR q.question_type = sqlc.arg(question_type)::text)
  AND (sqlc.arg(target_level)::text = '' OR q.target_level = sqlc.arg(target_level)::text)
  AND (sqlc.arg(difficulty_filter)::text = '' OR q.difficulty = sqlc.arg(difficulty_filter)::cbt_question_difficulty_enum)
  AND (sqlc.arg(cognitive_level)::text = '' OR q.cognitive_level = sqlc.arg(cognitive_level)::text)
  AND (sqlc.arg(material_topic)::text = '' OR q.material_topic ILIKE '%' || sqlc.arg(material_topic)::text || '%')
  AND (
    sqlc.arg(metadata_filter)::text = ''
    OR (
      sqlc.arg(metadata_filter)::text = 'complete'
      AND NULLIF(btrim(COALESCE(q.target_level, '')), '') IS NOT NULL
      AND NULLIF(btrim(q.cp_ref), '') IS NOT NULL
      AND (NULLIF(btrim(q.tp_ref), '') IS NOT NULL OR NULLIF(btrim(q.kd_ref), '') IS NOT NULL)
      AND NULLIF(btrim(q.cognitive_level), '') IS NOT NULL
    )
    OR (
      sqlc.arg(metadata_filter)::text = 'gap'
      AND (
        NULLIF(btrim(COALESCE(q.target_level, '')), '') IS NULL
        OR NULLIF(btrim(q.cp_ref), '') IS NULL
        OR (NULLIF(btrim(q.tp_ref), '') IS NULL AND NULLIF(btrim(q.kd_ref), '') IS NULL)
        OR NULLIF(btrim(q.cognitive_level), '') IS NULL
      )
    )
  )
  AND (sqlc.arg(hots_filter)::text = '' OR (sqlc.arg(hots_filter)::text = 'yes' AND q.hots_flag = TRUE) OR (sqlc.arg(hots_filter)::text = 'no' AND q.hots_flag = FALSE))
  AND (
    sqlc.arg(is_admin)::bool
    OR q.status = 'published'
    OR (sqlc.arg(can_use_in_package)::bool AND q.workflow_status IN ('approved', 'published'))
    OR q.author_username = sqlc.arg(actor_username)::text
    OR EXISTS (
      SELECT 1 FROM cbt_event_members m
      WHERE m.event_id = q.event_id
        AND m.user_id = sqlc.arg(actor_user_id)::uuid
        AND m.role IN ('reviewer', 'panitia')
        AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
    )
    OR EXISTS (
      SELECT 1 FROM bank_soal_reviewer_scopes rs
      WHERE rs.user_id = sqlc.arg(actor_user_id)::uuid
        AND (
          (rs.can_review = TRUE AND q.workflow_status IN ('submitted', 'review', 'revision_needed', 'reviewed'))
          OR (rs.can_approve = TRUE AND q.workflow_status IN ('reviewed', 'approved', 'published'))
        )
        AND (rs.subject_id IS NULL OR rs.subject_id = q.subject_id)
        AND (
          rs.grade_level IS NULL
          OR rs.grade_level = CASE q.target_level WHEN 'VII' THEN 7 WHEN 'VIII' THEN 8 WHEN 'IX' THEN 9 ELSE NULL END
        )
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
    OR q.tp_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.kd_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR q.indicator_ref ILIKE '%' || sqlc.arg(search_query)::text || '%'
  );

-- name: GetCbtQuestion :one
SELECT q.id, q.event_id, q.subject_id, q.code, q.question_text, q.question_type, q.options,
       q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
       q.answer_key, q.explanation, q.difficulty, q.status, q.created_at, q.updated_at,
       q.stem_html, q.stem_latex, q.stimulus_html, q.stimulus_latex,
       q.explanation_html, q.rubric_html,
       q.academic_phase, q.target_level,
       q.cp_ref, q.tp_ref, q.kd_ref, q.indicator_ref,
       q.material_topic, q.cognitive_level, q.hots_flag,
       q.media_asset_ids, q.workflow_status, q.version,
       q.version_group_id, q.version_number, q.source_question_id,
       q.supersedes_question_id, q.is_latest_version, q.version_note,
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
       q.academic_phase, q.target_level,
       q.cp_ref, q.tp_ref, q.kd_ref, q.indicator_ref,
       q.material_topic, q.cognitive_level, q.hots_flag,
       q.media_asset_ids, q.workflow_status, q.version,
       q.version_group_id, q.version_number, q.source_question_id,
       q.supersedes_question_id, q.is_latest_version, q.version_note,
       q.author_username,
       COALESCE(NULLIF(btrim(author_emp.nama), ''), NULLIF(btrim(author_user.display_name), ''), q.author_username) AS author_display_name,
       q.reviewer_username,
       COALESCE(NULLIF(btrim(reviewer_emp.nama), ''), NULLIF(btrim(reviewer_user.display_name), ''), q.reviewer_username) AS reviewer_display_name,
       q.reviewed_at,
       q.approver_username,
       COALESCE(NULLIF(btrim(approver_emp.nama), ''), NULLIF(btrim(approver_user.display_name), ''), q.approver_username) AS approver_display_name,
       q.approved_at, q.writer_notes, q.review_notes,
       COALESCE(pkg_usage.package_count, 0)::int AS package_count,
       COALESCE(answer_usage.answer_count, 0)::int AS answer_count
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
LEFT JOIN users author_user ON author_user.username = q.author_username
LEFT JOIN employees author_emp ON author_emp.id = author_user.employee_id
LEFT JOIN users reviewer_user ON reviewer_user.username = q.reviewer_username
LEFT JOIN employees reviewer_emp ON reviewer_emp.id = reviewer_user.employee_id
LEFT JOIN users approver_user ON approver_user.username = q.approver_username
LEFT JOIN employees approver_emp ON approver_emp.id = approver_user.employee_id
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
WITH new_question AS (
  SELECT gen_random_uuid() AS id
)
INSERT INTO cbt_questions (
  id, event_id, subject_id, code, question_text, question_type, options,
  option_a, option_b, option_c, option_d, option_e,
  answer_key, explanation, difficulty, status,
  stem_html, stem_latex, stimulus_html, stimulus_latex,
  explanation_html, rubric_html,
  academic_phase, target_level,
  cp_ref, tp_ref, kd_ref, indicator_ref,
  material_topic, cognitive_level, hots_flag,
  media_asset_ids, workflow_status, version,
  version_group_id, version_number, source_question_id,
  supersedes_question_id, is_latest_version, version_note,
  author_username, reviewer_username, reviewed_at,
  approver_username, approved_at, writer_notes, review_notes
)
SELECT
  new_question.id, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
  $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
  $31, $32, $33,
  COALESCE(sqlc.narg(version_group_id)::uuid, new_question.id),
  sqlc.arg(version_number),
  sqlc.narg(source_question_id)::uuid,
  sqlc.narg(supersedes_question_id)::uuid,
  sqlc.arg(is_latest_version),
  sqlc.arg(version_note),
  $34, $35, $36, $37, $38, $39, $40
FROM new_question
RETURNING *;

-- name: GetNextCbtQuestionVersionNumber :one
SELECT (COALESCE(MAX(version_number), 0)::int + 1)::int
FROM cbt_questions
WHERE version_group_id = $1;

-- name: MarkCbtQuestionVersionNotLatest :exec
UPDATE cbt_questions
SET is_latest_version = FALSE,
    updated_at = NOW()
WHERE id = $1;

-- name: MarkCbtQuestionVersionGroupNotLatest :exec
UPDATE cbt_questions
SET is_latest_version = FALSE,
    updated_at = NOW()
WHERE version_group_id = $1
  AND is_latest_version = TRUE;

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
  target_level       = $24,
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

-- name: CreateBankSoalQuestionWorkflowEvent :one
INSERT INTO bank_soal_question_workflow_events (
  question_id,
  actor_user_id,
  actor_username,
  from_status,
  to_status,
  action,
  note,
  metadata
)
VALUES (
  sqlc.arg(question_id),
  sqlc.narg(actor_user_id),
  sqlc.arg(actor_username),
  sqlc.arg(from_status),
  sqlc.arg(to_status),
  sqlc.arg(action),
  sqlc.arg(note),
  sqlc.arg(metadata)
)
RETURNING *;

-- name: ListCbtQuestionTimeline :many
SELECT log.id,
       log.question_id,
       log.actor_username,
       COALESCE(NULLIF(btrim(actor_emp.nama), ''), NULLIF(btrim(actor_user.display_name), ''), log.actor_username) AS actor_display_name,
       log.action,
       log.note,
       log.metadata,
       log.created_at
FROM cbt_question_audit_logs log
LEFT JOIN users actor_user ON actor_user.username = log.actor_username
LEFT JOIN employees actor_emp ON actor_emp.id = actor_user.employee_id
WHERE log.question_id = $1
ORDER BY log.created_at ASC, log.id ASC;

-- name: ListCbtQuestionVersions :many
SELECT q.id, q.code, q.workflow_status, q.status, q.version_number,
       q.is_latest_version, q.source_question_id, q.supersedes_question_id,
       q.version_note, q.created_at, q.updated_at,
       q.author_username,
       COALESCE(NULLIF(btrim(author_emp.nama), ''), NULLIF(btrim(author_user.display_name), ''), q.author_username) AS author_display_name,
       q.reviewer_username,
       COALESCE(NULLIF(btrim(reviewer_emp.nama), ''), NULLIF(btrim(reviewer_user.display_name), ''), q.reviewer_username) AS reviewer_display_name,
       q.approver_username,
       COALESCE(NULLIF(btrim(approver_emp.nama), ''), NULLIF(btrim(approver_user.display_name), ''), q.approver_username) AS approver_display_name
FROM cbt_questions q
LEFT JOIN users author_user ON author_user.username = q.author_username
LEFT JOIN employees author_emp ON author_emp.id = author_user.employee_id
LEFT JOIN users reviewer_user ON reviewer_user.username = q.reviewer_username
LEFT JOIN employees reviewer_emp ON reviewer_emp.id = reviewer_user.employee_id
LEFT JOIN users approver_user ON approver_user.username = q.approver_username
LEFT JOIN employees approver_emp ON approver_emp.id = approver_user.employee_id
WHERE q.version_group_id = (
  SELECT source.version_group_id FROM cbt_questions source WHERE source.id = $1
)
ORDER BY q.version_number DESC, q.created_at DESC;

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


-- name: GetCbtQuestionSummaryCounts :one
SELECT
  COUNT(*)::bigint AS total,
  COUNT(*) FILTER (WHERE q.status = 'draft')::bigint AS draft,
  COUNT(*) FILTER (WHERE q.workflow_status IN ('review', 'submitted'))::bigint AS review,
  COUNT(*) FILTER (WHERE q.workflow_status = 'rejected')::bigint AS rejected,
  COUNT(*) FILTER (WHERE q.workflow_status = 'approved')::bigint AS approved,
  COUNT(*) FILTER (WHERE q.status = 'published')::bigint AS published,
  COALESCE(SUM(pkg_usage.package_count), 0)::bigint AS package_usage
FROM cbt_questions q
LEFT JOIN LATERAL (
  SELECT COUNT(*)::bigint AS package_count
  FROM cbt_package_questions pq
  WHERE pq.question_id = q.id
) pkg_usage ON TRUE
WHERE TRUE
  AND (
    sqlc.arg(is_admin)::bool
    OR q.status = 'published'
    OR (sqlc.arg(can_use_in_package)::bool AND q.workflow_status IN ('approved', 'published'))
    OR q.author_username = sqlc.arg(actor_username)::text
    OR EXISTS (
      SELECT 1 FROM cbt_event_members m
      WHERE m.event_id = q.event_id
        AND m.user_id = sqlc.arg(actor_user_id)::uuid
        AND m.role IN ('reviewer', 'panitia')
        AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
    )
    OR EXISTS (
      SELECT 1 FROM bank_soal_reviewer_scopes rs
      WHERE rs.user_id = sqlc.arg(actor_user_id)::uuid
        AND (
          (rs.can_review = TRUE AND q.workflow_status IN ('submitted', 'review', 'revision_needed', 'reviewed'))
          OR (rs.can_approve = TRUE AND q.workflow_status IN ('reviewed', 'approved', 'published'))
        )
        AND (rs.subject_id IS NULL OR rs.subject_id = q.subject_id)
        AND (
          rs.grade_level IS NULL
          OR rs.grade_level = CASE q.target_level WHEN 'VII' THEN 7 WHEN 'VIII' THEN 8 WHEN 'IX' THEN 9 ELSE NULL END
        )
    )
  );

-- name: ListCbtQuestionSummaryBySubject :many
SELECT q.subject_id, s.name AS subject_name, s.code AS subject_code, COUNT(*)::bigint AS total
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
WHERE TRUE
  AND (
    sqlc.arg(is_admin)::bool
    OR q.status = 'published'
    OR (sqlc.arg(can_use_in_package)::bool AND q.workflow_status IN ('approved', 'published'))
    OR q.author_username = sqlc.arg(actor_username)::text
    OR EXISTS (
      SELECT 1 FROM cbt_event_members m
      WHERE m.event_id = q.event_id
        AND m.user_id = sqlc.arg(actor_user_id)::uuid
        AND m.role IN ('reviewer', 'panitia')
        AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
    )
    OR EXISTS (
      SELECT 1 FROM bank_soal_reviewer_scopes rs
      WHERE rs.user_id = sqlc.arg(actor_user_id)::uuid
        AND (
          (rs.can_review = TRUE AND q.workflow_status IN ('submitted', 'review', 'revision_needed', 'reviewed'))
          OR (rs.can_approve = TRUE AND q.workflow_status IN ('reviewed', 'approved', 'published'))
        )
        AND (rs.subject_id IS NULL OR rs.subject_id = q.subject_id)
        AND (
          rs.grade_level IS NULL
          OR rs.grade_level = CASE q.target_level WHEN 'VII' THEN 7 WHEN 'VIII' THEN 8 WHEN 'IX' THEN 9 ELSE NULL END
        )
    )
  )
GROUP BY q.subject_id, s.name, s.code
ORDER BY total DESC, s.name ASC
LIMIT 8;

-- name: ListCbtQuestionSummaryByCognitiveLevel :many
SELECT COALESCE(NULLIF(btrim(q.cognitive_level), ''), 'Belum diisi') AS cognitive_level, COUNT(*)::bigint AS total
FROM cbt_questions q
WHERE TRUE
  AND (
    sqlc.arg(is_admin)::bool
    OR q.status = 'published'
    OR (sqlc.arg(can_use_in_package)::bool AND q.workflow_status IN ('approved', 'published'))
    OR q.author_username = sqlc.arg(actor_username)::text
    OR EXISTS (
      SELECT 1 FROM cbt_event_members m
      WHERE m.event_id = q.event_id
        AND m.user_id = sqlc.arg(actor_user_id)::uuid
        AND m.role IN ('reviewer', 'panitia')
        AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
    )
    OR EXISTS (
      SELECT 1 FROM bank_soal_reviewer_scopes rs
      WHERE rs.user_id = sqlc.arg(actor_user_id)::uuid
        AND (
          (rs.can_review = TRUE AND q.workflow_status IN ('submitted', 'review', 'revision_needed', 'reviewed'))
          OR (rs.can_approve = TRUE AND q.workflow_status IN ('reviewed', 'approved', 'published'))
        )
        AND (rs.subject_id IS NULL OR rs.subject_id = q.subject_id)
        AND (
          rs.grade_level IS NULL
          OR rs.grade_level = CASE q.target_level WHEN 'VII' THEN 7 WHEN 'VIII' THEN 8 WHEN 'IX' THEN 9 ELSE NULL END
        )
    )
  )
GROUP BY COALESCE(NULLIF(btrim(q.cognitive_level), ''), 'Belum diisi')
ORDER BY total DESC, cognitive_level ASC;

-- name: ListCbtQuestionSummaryRecent :many
SELECT q.id, q.code, q.subject_id, s.name AS subject_name, s.code AS subject_code,
       q.material_topic, q.workflow_status, q.status, q.author_username,
       q.reviewer_username, q.created_at, q.updated_at
FROM cbt_questions q
JOIN subjects s ON s.id = q.subject_id
WHERE TRUE
  AND (
    sqlc.arg(is_admin)::bool
    OR q.status = 'published'
    OR (sqlc.arg(can_use_in_package)::bool AND q.workflow_status IN ('approved', 'published'))
    OR q.author_username = sqlc.arg(actor_username)::text
    OR EXISTS (
      SELECT 1 FROM cbt_event_members m
      WHERE m.event_id = q.event_id
        AND m.user_id = sqlc.arg(actor_user_id)::uuid
        AND m.role IN ('reviewer', 'panitia')
        AND (m.subject_id IS NULL OR m.subject_id = q.subject_id)
    )
    OR EXISTS (
      SELECT 1 FROM bank_soal_reviewer_scopes rs
      WHERE rs.user_id = sqlc.arg(actor_user_id)::uuid
        AND (
          (rs.can_review = TRUE AND q.workflow_status IN ('submitted', 'review', 'revision_needed', 'reviewed'))
          OR (rs.can_approve = TRUE AND q.workflow_status IN ('reviewed', 'approved', 'published'))
        )
        AND (rs.subject_id IS NULL OR rs.subject_id = q.subject_id)
        AND (
          rs.grade_level IS NULL
          OR rs.grade_level = CASE q.target_level WHEN 'VII' THEN 7 WHEN 'VIII' THEN 8 WHEN 'IX' THEN 9 ELSE NULL END
        )
    )
  )
ORDER BY q.updated_at DESC, q.created_at DESC
LIMIT 5;
