-- name: ListCbtPackages :many
SELECT p.id, p.event_id, p.subject_id, s.name AS subject_name, s.code AS subject_code,
       p.title, p.description, p.duration_minutes, p.randomize_questions,
       COALESCE(p.randomize_options, FALSE)::boolean AS randomize_options,
       COALESCE(p.source_mode, 'teacher_class')::text AS source_mode,
       COALESCE(p.draw_pg_count, 0)::int AS draw_pg_count,
       COALESCE(p.draw_essay_count, 0)::int AS draw_essay_count,
       COALESCE(p.random_seed, '')::text AS random_seed,
       COALESCE(p.composition_log, '{}'::jsonb) AS composition_log,
       p.is_active,
       p.created_at, p.updated_at,
       COUNT(pq.question_id)::int AS question_count
FROM cbt_packages p
JOIN subjects s ON s.id = p.subject_id
LEFT JOIN cbt_package_questions pq ON pq.package_id = p.id
WHERE (sqlc.arg(event_id)::uuid IS NULL OR p.event_id = sqlc.arg(event_id)::uuid)
GROUP BY p.id, s.name, s.code
ORDER BY p.created_at DESC;

-- name: CreateCbtPackage :one
INSERT INTO cbt_packages (id, event_id, subject_id, title, description, duration_minutes, randomize_questions, is_active, source_mode, randomize_options, draw_pg_count, draw_essay_count, random_seed, composition_log)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: DeleteCbtPackage :execrows
DELETE FROM cbt_packages WHERE id = $1;

-- name: GetCbtPackageUsage :one
SELECT COUNT(s.id)::int AS session_count
FROM cbt_packages p
LEFT JOIN cbt_exam_sessions s ON s.package_id = p.id
WHERE p.id = $1
GROUP BY p.id;

-- name: AddCbtPackageQuestion :exec
INSERT INTO cbt_package_questions (package_id, question_id, position, points)
VALUES ($1, $2, $3, $4);

-- name: ListCbtPackageQuestions :many
SELECT pq.package_id, pq.question_id, pq.position, pq.points,
       q.event_id, q.code AS question_code, q.question_text, q.question_type, q.difficulty, q.status, q.workflow_status,
       q.cp_ref, q.tp_ref, q.kd_ref, q.material_topic, q.cognitive_level, q.hots_flag
FROM cbt_package_questions pq
JOIN cbt_questions q ON q.id = pq.question_id
JOIN cbt_packages p ON p.id = pq.package_id
WHERE (sqlc.arg(event_id)::uuid IS NULL OR p.event_id = sqlc.arg(event_id)::uuid)
ORDER BY pq.package_id, pq.position ASC;

-- name: ListCbtEventPackages :many
SELECT p.id, p.event_id, p.subject_id, s.name AS subject_name, s.code AS subject_code,
       p.title, p.description, p.duration_minutes, p.randomize_questions,
       COALESCE(p.randomize_options, FALSE)::boolean AS randomize_options,
       COALESCE(p.source_mode, 'teacher_class')::text AS source_mode,
       COALESCE(p.draw_pg_count, 0)::int AS draw_pg_count,
       COALESCE(p.draw_essay_count, 0)::int AS draw_essay_count,
       COALESCE(p.random_seed, '')::text AS random_seed,
       COALESCE(p.composition_log, '{}'::jsonb) AS composition_log,
       p.is_active,
       p.created_at, p.updated_at,
       COUNT(pq.question_id)::int AS question_count,
       COUNT(pq.question_id) FILTER (WHERE q.status = 'published')::int AS published_question_count,
       COUNT(pq.question_id) FILTER (WHERE q.event_id = p.event_id)::int AS event_question_count
FROM cbt_packages p
JOIN subjects s ON s.id = p.subject_id
LEFT JOIN cbt_package_questions pq ON pq.package_id = p.id
LEFT JOIN cbt_questions q ON q.id = pq.question_id
WHERE p.event_id = $1
GROUP BY p.id, s.name, s.code
ORDER BY s.name ASC, p.created_at DESC;

-- name: GetCbtPackageQuestionQuality :one
SELECT
  p.is_active,
  COALESCE(p.randomize_options, FALSE)::boolean AS randomize_options,
  COALESCE(p.source_mode, 'teacher_class')::text AS source_mode,
  COALESCE(p.draw_pg_count, 0)::int AS draw_pg_count,
  COALESCE(p.draw_essay_count, 0)::int AS draw_essay_count,
  COUNT(q.id)::int AS total_questions,
  COUNT(q.id) FILTER (WHERE q.status = 'published')::int AS published_questions,
  COUNT(q.id) FILTER (WHERE q.status <> 'published')::int AS unpublished_questions,
  COUNT(q.id) FILTER (
    WHERE q.cp_ref = ''
       OR (q.tp_ref = '' AND q.kd_ref = '')
       OR q.cognitive_level = ''
  )::int AS metadata_gap_questions
FROM cbt_packages p
LEFT JOIN cbt_package_questions pq ON pq.package_id = p.id
LEFT JOIN cbt_questions q ON q.id = pq.question_id
WHERE p.id = $1
GROUP BY p.id, p.is_active;

-- name: GetExamQuestions :many
SELECT
  q.id, q.code, q.question_text, q.question_type, q.options,
  q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
  q.stem_html, q.stem_latex, q.stimulus_html, q.stimulus_latex, q.media_asset_ids
FROM cbt_package_questions pq
JOIN cbt_questions q ON q.id = pq.question_id
WHERE pq.package_id = $1 AND q.status = 'published'
ORDER BY pq.position ASC;
