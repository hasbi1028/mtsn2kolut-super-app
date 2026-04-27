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
       q.code AS question_code, q.question_text
FROM cbt_package_questions pq
JOIN cbt_questions q ON q.id = pq.question_id
ORDER BY pq.package_id, pq.position ASC;
