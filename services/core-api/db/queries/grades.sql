-- name: ListGradeComponents :many
SELECT gc.id, gc.assignment_id, gc.title, gc.category, gc.weight, gc.max_score,
       gc.is_published, gc.created_at, gc.updated_at,
       csa.class_id, c.name AS class_name, c.code AS class_code,
       csa.subject_id, s.name AS subject_name, s.code AS subject_code,
       csa.teacher_employee_id, e.nama AS teacher_name
FROM grade_components gc
JOIN class_subject_assignments csa ON csa.id = gc.assignment_id
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects s ON s.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
WHERE (sqlc.arg(assignment_id)::uuid IS NULL OR gc.assignment_id = sqlc.arg(assignment_id)::uuid)
ORDER BY c.level ASC, c.name ASC, s.name ASC, gc.created_at DESC;

-- name: GetGradeComponent :one
SELECT gc.id, gc.assignment_id, gc.title, gc.category, gc.weight, gc.max_score,
       gc.is_published, gc.created_at, gc.updated_at
FROM grade_components gc
WHERE gc.id = $1;

-- name: CreateGradeComponent :one
INSERT INTO grade_components (assignment_id, title, category, weight, max_score, is_published)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: DeleteGradeComponent :exec
DELETE FROM grade_components WHERE id = $1;

-- name: ListGradebookSummary :many
WITH component_set AS (
  SELECT gc.id, gc.weight, gc.max_score
  FROM grade_components gc
  WHERE gc.assignment_id = $1
),
score_rows AS (
  SELECT st.id AS student_id,
         st.nis,
         st.nisn,
         st.nama,
         cs.id AS component_id,
         cs.weight,
         cs.max_score,
         ge.score
  FROM class_subject_assignments csa
  JOIN students st ON st.class_id = csa.class_id
  LEFT JOIN component_set cs ON TRUE
  LEFT JOIN grade_entries ge ON ge.component_id = cs.id AND ge.student_id = st.id
  WHERE csa.id = $1
    AND st.is_active = TRUE
)
SELECT student_id,
       nis,
       nisn,
       nama,
       COUNT(component_id)::int AS component_count,
       COUNT(score)::int AS filled_count,
       COALESCE(
         ROUND(
           (
             SUM(CASE WHEN score IS NOT NULL AND max_score > 0 THEN (score / max_score) * weight ELSE 0 END)
             / NULLIF(SUM(CASE WHEN score IS NOT NULL THEN weight ELSE 0 END), 0)
           ) * 100
         )::numeric,
         -1
       )::double precision AS final_score
FROM score_rows
GROUP BY student_id, nis, nisn, nama
ORDER BY nama ASC;

-- name: ListGradeEntriesByComponent :many
SELECT st.id AS student_id,
       st.nis,
       st.nisn,
       st.nama,
       ge.id AS entry_id,
       COALESCE(ge.score, -1)::double precision AS score,
       ge.notes,
       ge.graded_by,
       ge.graded_at
FROM grade_components gc
JOIN class_subject_assignments csa ON csa.id = gc.assignment_id
JOIN students st ON st.class_id = csa.class_id
LEFT JOIN grade_entries ge ON ge.component_id = gc.id AND ge.student_id = st.id
WHERE gc.id = $1
  AND st.is_active = TRUE
ORDER BY st.nama ASC;

-- name: UpsertGradeEntry :one
INSERT INTO grade_entries (component_id, student_id, score, notes, graded_by, graded_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (component_id, student_id) DO UPDATE
  SET score      = EXCLUDED.score,
      notes      = EXCLUDED.notes,
      graded_by  = EXCLUDED.graded_by,
      graded_at  = NOW(),
      updated_at = NOW()
RETURNING *;
