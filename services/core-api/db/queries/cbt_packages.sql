-- name: ListCbtPackages :many
SELECT p.id, p.subject_id, s.name AS subject_name, s.code AS subject_code,
       p.title, p.description, p.duration_minutes, p.randomize_questions, p.is_active,
       p.created_at, p.updated_at,
       COUNT(pq.question_id)::int AS question_count
FROM cbt_packages p
JOIN subjects s ON s.id = p.subject_id
LEFT JOIN cbt_package_questions pq ON pq.package_id = p.id
GROUP BY p.id, s.name, s.code
ORDER BY p.created_at DESC;

-- name: CreateCbtPackage :one
INSERT INTO cbt_packages (id, subject_id, title, description, duration_minutes, randomize_questions, is_active)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: DeleteCbtPackage :exec
DELETE FROM cbt_packages WHERE id = $1;

-- name: AddCbtPackageQuestion :exec
INSERT INTO cbt_package_questions (package_id, question_id, position, points)
VALUES ($1, $2, $3, $4);

-- name: ListCbtPackageQuestions :many
SELECT pq.package_id, pq.question_id, pq.position, pq.points,
       q.code AS question_code, q.question_text, q.question_type, q.difficulty, q.status, q.workflow_status,
       q.cp_ref, q.tp_ref, q.kd_ref, q.material_topic, q.cognitive_level, q.hots_flag
FROM cbt_package_questions pq
JOIN cbt_questions q ON q.id = pq.question_id
ORDER BY pq.package_id, pq.position ASC;

-- name: GetExamQuestions :many
SELECT
  q.id, q.code, q.question_text, q.question_type, q.options,
  q.option_a, q.option_b, q.option_c, q.option_d, q.option_e,
  q.stem_html, q.stem_latex, q.stimulus_html, q.stimulus_latex, q.media_asset_ids
FROM cbt_package_questions pq
JOIN cbt_questions q ON q.id = pq.question_id
WHERE pq.package_id = $1 AND q.status = 'published'
ORDER BY pq.position ASC;
