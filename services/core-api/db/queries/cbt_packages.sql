-- name: ListCbtPackages :many
SELECT p.id, p.event_id, p.subject_id, s.name AS subject_name, s.code AS subject_code,
       p.title, p.description, p.duration_minutes, p.randomize_questions,
       COALESCE(p.randomize_options, FALSE)::boolean AS randomize_options,
       COALESCE(p.source_mode, 'teacher_class')::text AS source_mode,
       COALESCE(p.draw_pg_count, 0)::int AS draw_pg_count,
       COALESCE(p.draw_essay_count, 0)::int AS draw_essay_count,
       COALESCE(p.random_seed, '')::text AS random_seed,
       COALESCE(p.composition_log, '{}'::jsonb) AS composition_log,
       p.locked_at, p.locked_by, p.lock_reason, p.snapshot_version,
       p.is_active,
       p.created_at, p.updated_at,
       COUNT(DISTINCT pq.question_id)::int AS question_count,
       COUNT(DISTINCT ses.id)::int AS session_count
FROM cbt_packages p
JOIN subjects s ON s.id = p.subject_id
LEFT JOIN cbt_package_questions pq ON pq.package_id = p.id
LEFT JOIN cbt_exam_sessions ses ON ses.package_id = p.id
WHERE (sqlc.arg(event_id)::uuid IS NULL OR p.event_id = sqlc.arg(event_id)::uuid)
GROUP BY p.id, s.name, s.code
ORDER BY p.created_at DESC;

-- name: GetCbtPackageDetail :one
SELECT p.id, p.event_id, e.title AS event_title, e.status AS event_status,
       p.subject_id, s.name AS subject_name, s.code AS subject_code,
       p.title, p.description, p.duration_minutes, p.randomize_questions,
       COALESCE(p.randomize_options, FALSE)::boolean AS randomize_options,
       COALESCE(p.source_mode, 'teacher_class')::text AS source_mode,
       COALESCE(p.draw_pg_count, 0)::int AS draw_pg_count,
       COALESCE(p.draw_essay_count, 0)::int AS draw_essay_count,
       COALESCE(p.random_seed, '')::text AS random_seed,
       COALESCE(p.composition_log, '{}'::jsonb) AS composition_log,
       p.locked_at, p.locked_by, p.lock_reason, p.snapshot_version,
       p.is_active,
       p.created_at, p.updated_at,
       COUNT(DISTINCT pq.question_id)::int AS question_count,
       COUNT(DISTINCT ses.id)::int AS session_count
FROM cbt_packages p
JOIN subjects s ON s.id = p.subject_id
LEFT JOIN cbt_exam_events e ON e.id = p.event_id
LEFT JOIN cbt_package_questions pq ON pq.package_id = p.id
LEFT JOIN cbt_exam_sessions ses ON ses.package_id = p.id
WHERE p.id = $1
GROUP BY p.id, e.title, e.status, s.name, s.code;

-- name: LockCbtPackageForEdit :one
SELECT id, event_id, subject_id, title, description, duration_minutes,
       randomize_questions, COALESCE(randomize_options, FALSE)::boolean AS randomize_options,
       COALESCE(source_mode, 'teacher_class')::text AS source_mode,
       COALESCE(draw_pg_count, 0)::int AS draw_pg_count,
       COALESCE(draw_essay_count, 0)::int AS draw_essay_count,
       COALESCE(random_seed, '')::text AS random_seed,
       COALESCE(composition_log, '{}'::jsonb) AS composition_log,
       locked_at, locked_by, lock_reason, snapshot_version,
       is_active, created_at, updated_at
FROM cbt_packages
WHERE id = $1
FOR UPDATE;

-- name: UpdateCbtPackageMetadata :one
UPDATE cbt_packages
SET title = sqlc.arg(title),
    description = sqlc.arg(description),
    duration_minutes = sqlc.arg(duration_minutes),
    randomize_questions = sqlc.arg(randomize_questions),
    randomize_options = sqlc.arg(randomize_options),
    source_mode = sqlc.arg(source_mode),
    draw_pg_count = sqlc.arg(draw_pg_count),
    draw_essay_count = sqlc.arg(draw_essay_count),
    random_seed = sqlc.arg(random_seed),
    is_active = sqlc.arg(is_active),
    composition_log = jsonb_set(
      COALESCE(composition_log, '{}'::jsonb),
      '{metadata_updated_at}',
      to_jsonb(NOW()::text),
      TRUE
    ),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND locked_at IS NULL
RETURNING *;

-- name: CreateCbtPackage :one
INSERT INTO cbt_packages (id, event_id, subject_id, title, description, duration_minutes, randomize_questions, is_active, source_mode, randomize_options, draw_pg_count, draw_essay_count, random_seed, composition_log)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: DeleteCbtPackage :execrows
DELETE FROM cbt_packages WHERE id = $1 AND locked_at IS NULL;

-- name: GetCbtPackageUsage :one
SELECT COUNT(s.id)::int AS session_count
FROM cbt_packages p
LEFT JOIN cbt_exam_sessions s ON s.package_id = p.id
WHERE p.id = $1
GROUP BY p.id;

-- name: AddCbtPackageQuestion :exec
INSERT INTO cbt_package_questions (package_id, question_id, position, points)
SELECT $1, $2, $3, $4
WHERE NOT EXISTS (
  SELECT 1
  FROM cbt_packages p
  WHERE p.id = $1
    AND p.locked_at IS NOT NULL
);

-- name: DeleteCbtPackageQuestions :execrows
DELETE FROM cbt_package_questions
WHERE package_id = $1
  AND EXISTS (
    SELECT 1
    FROM cbt_packages p
    WHERE p.id = $1
      AND p.locked_at IS NULL
  );

-- name: ListCbtPackageQuestions :many
SELECT pq.package_id, pq.question_id, pq.position, pq.points,
       q.event_id, q.code AS question_code, q.question_text, q.question_type, q.difficulty, q.status, q.workflow_status,
       q.cp_ref, q.tp_ref, q.kd_ref, q.material_topic, q.cognitive_level, q.hots_flag
FROM cbt_package_questions pq
JOIN cbt_questions q ON q.id = pq.question_id
JOIN cbt_packages p ON p.id = pq.package_id
WHERE (sqlc.arg(event_id)::uuid IS NULL OR p.event_id = sqlc.arg(event_id)::uuid)
ORDER BY pq.package_id, pq.position ASC;

-- name: ListCbtPackageQuestionsByPackage :many
SELECT pq.package_id, pq.question_id, pq.position, pq.points,
       q.event_id, q.subject_id, q.code AS question_code, q.question_text, q.question_type, q.difficulty, q.status, q.workflow_status,
       q.cp_ref, q.tp_ref, q.kd_ref, q.material_topic, q.cognitive_level, q.hots_flag
FROM cbt_package_questions pq
JOIN cbt_questions q ON q.id = pq.question_id
WHERE pq.package_id = $1
ORDER BY pq.position ASC, pq.created_at ASC;

-- name: ListCbtEventPackages :many
SELECT p.id, p.event_id, p.subject_id, s.name AS subject_name, s.code AS subject_code,
       p.title, p.description, p.duration_minutes, p.randomize_questions,
       COALESCE(p.randomize_options, FALSE)::boolean AS randomize_options,
       COALESCE(p.source_mode, 'teacher_class')::text AS source_mode,
       COALESCE(p.draw_pg_count, 0)::int AS draw_pg_count,
       COALESCE(p.draw_essay_count, 0)::int AS draw_essay_count,
       COALESCE(p.random_seed, '')::text AS random_seed,
       COALESCE(p.composition_log, '{}'::jsonb) AS composition_log,
       p.locked_at, p.locked_by, p.lock_reason, p.snapshot_version,
       p.is_active,
       p.created_at, p.updated_at,
       COUNT(DISTINCT pq.question_id)::int AS question_count,
       COUNT(DISTINCT pq.question_id) FILTER (WHERE q.status = 'published')::int AS published_question_count,
       COUNT(DISTINCT pq.question_id) FILTER (WHERE q.event_id = p.event_id)::int AS event_question_count,
       COUNT(DISTINCT ses.id)::int AS session_count
FROM cbt_packages p
JOIN subjects s ON s.id = p.subject_id
LEFT JOIN cbt_package_questions pq ON pq.package_id = p.id
LEFT JOIN cbt_questions q ON q.id = pq.question_id
LEFT JOIN cbt_exam_sessions ses ON ses.package_id = p.id
WHERE p.event_id = $1
GROUP BY p.id, s.name, s.code
ORDER BY s.name ASC, p.created_at DESC;

-- name: CloneCbtPackage :one
INSERT INTO cbt_packages (
  id, event_id, subject_id, title, description, duration_minutes,
  randomize_questions, is_active, source_mode, randomize_options,
  draw_pg_count, draw_essay_count, random_seed, composition_log
)
SELECT gen_random_uuid(),
       event_id,
       subject_id,
       COALESCE(NULLIF(sqlc.arg(title)::text, ''), title || ' - Revisi'),
       description,
       duration_minutes,
       randomize_questions,
       is_active,
       COALESCE(source_mode, 'teacher_class'),
       COALESCE(randomize_options, FALSE),
       COALESCE(draw_pg_count, 0),
       COALESCE(draw_essay_count, 0),
       COALESCE(random_seed, ''),
       jsonb_set(
         COALESCE(composition_log, '{}'::jsonb),
         '{cloned_from_package_id}',
         to_jsonb(src.id::text),
         TRUE
       )
FROM cbt_packages src
WHERE src.id = sqlc.arg(source_id)
RETURNING *;

-- name: CloneCbtPackageQuestions :execrows
INSERT INTO cbt_package_questions (package_id, question_id, position, points)
SELECT sqlc.arg(target_id), question_id, position, points
FROM cbt_package_questions source_questions
WHERE source_questions.package_id = sqlc.arg(source_id)
ORDER BY position ASC;

-- name: ListCbtPackageReadiness :many
SELECT p.id, p.event_id, p.subject_id, s.name AS subject_name, s.code AS subject_code,
       p.title, p.duration_minutes, p.is_active, p.locked_at, p.snapshot_version,
       COUNT(DISTINCT pq.question_id)::int AS question_count,
       COUNT(DISTINCT pq.question_id) FILTER (WHERE q.question_type = 'multiple_choice')::int AS pg_count,
       COUNT(DISTINCT pq.question_id) FILTER (WHERE q.question_type = 'essay')::int AS essay_count,
       COALESCE(SUM(pq.points), 0)::int AS total_points,
       COUNT(DISTINCT pq.question_id) FILTER (WHERE q.status = 'published')::int AS published_count,
       COUNT(DISTINCT pq.question_id) FILTER (WHERE q.status <> 'published')::int AS unpublished_count,
       COUNT(DISTINCT pq.question_id) FILTER (
         WHERE q.cp_ref = ''
            OR (q.tp_ref = '' AND q.kd_ref = '')
            OR q.cognitive_level = ''
       )::int AS metadata_gap_count,
       COALESCE(usage.session_count, 0)::int AS session_count
FROM cbt_packages p
JOIN subjects s ON s.id = p.subject_id
LEFT JOIN cbt_package_questions pq ON pq.package_id = p.id
LEFT JOIN cbt_questions q ON q.id = pq.question_id
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS session_count
  FROM cbt_exam_sessions ses
  WHERE ses.package_id = p.id
) usage ON TRUE
WHERE (sqlc.arg(event_id)::uuid IS NULL OR p.event_id = sqlc.arg(event_id)::uuid)
GROUP BY p.id, s.name, s.code, usage.session_count
ORDER BY s.name ASC, p.title ASC;

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
WITH active_snapshot AS (
  SELECT snap.*
  FROM cbt_packages p
  JOIN cbt_package_question_snapshots snap
    ON snap.package_id = p.id
   AND snap.snapshot_version = p.snapshot_version
  WHERE p.id = $1
    AND p.snapshot_version > 0
),
has_snapshot AS (
  SELECT EXISTS(SELECT 1 FROM active_snapshot) AS available
)
SELECT
  rows.id, rows.code, rows.question_text, rows.question_type, rows.options,
  rows.option_a, rows.option_b, rows.option_c, rows.option_d, rows.option_e,
  rows.stem_html, rows.stem_latex, rows.stimulus_html, rows.stimulus_latex, rows.media_asset_ids
FROM (
  SELECT
    snap.position AS sort_position,
    snap.question_id AS id,
    snap.question_code AS code,
    snap.question_text,
    snap.question_type,
    snap.options,
    snap.option_a,
    snap.option_b,
    snap.option_c,
    snap.option_d,
    snap.option_e,
    snap.stem_html,
    snap.stem_latex,
    snap.stimulus_html,
    snap.stimulus_latex,
    snap.media_asset_ids
  FROM active_snapshot snap
  UNION ALL
  SELECT
    pq.position AS sort_position,
    q.id, q.code, q.question_text, q.question_type, q.options,
    q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
    q.stem_html, q.stem_latex, q.stimulus_html, q.stimulus_latex, q.media_asset_ids
  FROM cbt_package_questions pq
  JOIN cbt_questions q ON q.id = pq.question_id
  CROSS JOIN has_snapshot
  WHERE pq.package_id = $1
    AND q.status = 'published'
    AND has_snapshot.available = FALSE
) rows
ORDER BY rows.sort_position ASC;

-- name: GetCbtPackageLockState :one
SELECT id, locked_at, locked_by, lock_reason, snapshot_version
FROM cbt_packages
WHERE id = $1;

-- name: LockCbtPackageForSnapshot :one
UPDATE cbt_packages
SET locked_at = COALESCE(locked_at, NOW()),
    locked_by = CASE WHEN locked_at IS NULL THEN sqlc.arg(locked_by)::uuid ELSE locked_by END,
    lock_reason = CASE
      WHEN locked_at IS NULL THEN COALESCE(NULLIF(sqlc.arg(lock_reason)::text, ''), 'session_scheduled')
      ELSE lock_reason
    END,
    snapshot_version = CASE WHEN snapshot_version <= 0 THEN 1 ELSE snapshot_version END,
    updated_at = NOW()
WHERE id = sqlc.arg(package_id)
RETURNING id, locked_at, locked_by, lock_reason, snapshot_version;

-- name: CreateCbtPackageQuestionSnapshots :execrows
INSERT INTO cbt_package_question_snapshots (
  package_id, snapshot_version, question_id, position, points, question_code,
  question_text, question_type, options, option_a, option_b, option_c, option_d, option_e,
  answer_key, stem_html, stem_latex, stimulus_html, stimulus_latex,
  rubric_html, explanation_html, media_asset_ids, metadata
)
SELECT
  p.id,
  p.snapshot_version,
  q.id,
  pq.position,
  pq.points,
  q.code,
  q.question_text,
  q.question_type,
  COALESCE(q.options, '[]'::jsonb),
  q.option_a,
  q.option_b,
  q.option_c,
  q.option_d,
  q.option_e,
  q.answer_key,
  q.stem_html,
  q.stem_latex,
  q.stimulus_html,
  q.stimulus_latex,
  q.rubric_html,
  q.explanation_html,
  COALESCE(q.media_asset_ids, '[]'::jsonb),
  jsonb_build_object(
    'difficulty', q.difficulty,
    'academic_phase', q.academic_phase,
    'grade_level', q.grade_level,
    'cp_ref', q.cp_ref,
    'tp_ref', q.tp_ref,
    'kd_ref', q.kd_ref,
    'indicator_ref', q.indicator_ref,
    'material_topic', q.material_topic,
    'cognitive_level', q.cognitive_level,
    'hots_flag', q.hots_flag,
    'workflow_status', q.workflow_status,
    'version', q.version
  )
FROM cbt_packages p
JOIN cbt_package_questions pq ON pq.package_id = p.id
JOIN cbt_questions q ON q.id = pq.question_id
WHERE p.id = sqlc.arg(package_id)
  AND p.snapshot_version > 0
ON CONFLICT (package_id, snapshot_version, question_id) DO NOTHING;
