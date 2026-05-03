-- name: ListGradeComponents :many
SELECT gc.id, gc.assignment_id, gc.title, gc.category, gc.weight, gc.max_score,
       gc.is_published, gc.created_at, gc.updated_at,
       csa.class_id, c.name AS class_name, c.code AS class_code,
       csa.subject_id, s.name AS subject_name, s.code AS subject_code,
       csa.teacher_employee_id, e.nama AS teacher_name,
       nta.id AS source_non_test_assessment_id,
       COALESCE(nta.title, '') AS source_non_test_title,
       COALESCE(nta.assessment_type, '') AS source_non_test_type,
       nta.grade_synced_at AS source_non_test_synced_at,
       COALESCE(nta.grade_synced_by, '') AS source_non_test_synced_by
FROM grade_components gc
JOIN class_subject_assignments csa ON csa.id = gc.assignment_id
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects s ON s.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
LEFT JOIN LATERAL (
  SELECT id, title, assessment_type, grade_synced_at, grade_synced_by
  FROM non_test_assessments
  WHERE grade_component_id = gc.id
  ORDER BY grade_synced_at DESC NULLS LAST, updated_at DESC
  LIMIT 1
) nta ON TRUE
WHERE (sqlc.arg(assignment_id)::uuid IS NULL OR gc.assignment_id = sqlc.arg(assignment_id)::uuid)
  AND (NOT sqlc.arg(published_only)::boolean OR gc.is_published = TRUE)
ORDER BY c.level ASC, c.name ASC, s.name ASC, gc.created_at DESC;

-- name: ListGradeAssignmentStatuses :many
WITH component_rollup AS (
  SELECT csa.id AS assignment_id,
         COUNT(gc.id)::int AS component_count,
         COUNT(gc.id) FILTER (WHERE gc.is_published = TRUE)::int AS published_component_count,
         COUNT(gc.id) FILTER (WHERE gc.is_published = FALSE)::int AS draft_component_count
  FROM class_subject_assignments csa
  LEFT JOIN grade_components gc ON gc.assignment_id = csa.id
  GROUP BY csa.id
),
student_component_rollup AS (
  SELECT csa.id AS assignment_id,
         st.id AS student_id,
         COALESCE(cr.component_count, 0)::int AS component_count,
         COUNT(ge.score)::int AS filled_count
  FROM class_subject_assignments csa
  JOIN students st ON st.class_id = csa.class_id
  LEFT JOIN component_rollup cr ON cr.assignment_id = csa.id
  LEFT JOIN grade_components gc ON gc.assignment_id = csa.id
  LEFT JOIN grade_entries ge ON ge.component_id = gc.id AND ge.student_id = st.id
  WHERE st.is_active = TRUE
  GROUP BY csa.id, st.id, cr.component_count
)
SELECT csa.id AS assignment_id,
       csa.class_id,
       c.name AS class_name,
       c.code AS class_code,
       csa.subject_id,
       s.name AS subject_name,
       s.code AS subject_code,
       csa.teacher_employee_id,
       e.nama AS teacher_name,
       COALESCE(cr.component_count, 0)::int AS component_count,
       COALESCE(cr.published_component_count, 0)::int AS published_component_count,
       COALESCE(cr.draft_component_count, 0)::int AS draft_component_count,
       COUNT(scr.student_id)::int AS student_count,
       COUNT(scr.student_id) FILTER (
         WHERE scr.component_count > 0
           AND scr.filled_count = scr.component_count
       )::int AS ready_student_count,
       COUNT(scr.student_id) FILTER (
         WHERE scr.component_count = 0
           OR scr.filled_count < scr.component_count
       )::int AS incomplete_student_count,
       COALESCE(SUM(GREATEST(scr.component_count - scr.filled_count, 0)), 0)::int AS missing_grade_count,
       (gaf.assignment_id IS NOT NULL)::boolean AS is_finalized,
       gaf.finalized_by,
       gaf.notes,
       gaf.finalized_at
FROM class_subject_assignments csa
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects s ON s.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
LEFT JOIN component_rollup cr ON cr.assignment_id = csa.id
LEFT JOIN student_component_rollup scr ON scr.assignment_id = csa.id
LEFT JOIN grade_assignment_finalizations gaf ON gaf.assignment_id = csa.id
GROUP BY csa.id, csa.class_id, c.name, c.code, c.level,
         csa.subject_id, s.name, s.code,
         csa.teacher_employee_id, e.nama,
         cr.component_count, cr.published_component_count, cr.draft_component_count,
         gaf.assignment_id, gaf.finalized_by, gaf.notes, gaf.finalized_at
ORDER BY c.level ASC, c.name ASC, s.name ASC;

-- name: GetGradeComponent :one
SELECT gc.id, gc.assignment_id, gc.title, gc.category, gc.weight, gc.max_score,
       gc.is_published, gc.created_at, gc.updated_at
FROM grade_components gc
WHERE gc.id = $1;

-- name: GetNonTestGradeComponentSource :one
SELECT id
FROM non_test_assessments
WHERE grade_component_id = $1
ORDER BY grade_synced_at DESC NULLS LAST, updated_at DESC
LIMIT 1;

-- name: GetGradeComponentHighestScore :one
SELECT COALESCE(MAX(ge.score), -1)::double precision AS max_score
FROM grade_entries ge
WHERE ge.component_id = $1;

-- name: CreateGradeComponent :one
INSERT INTO grade_components (assignment_id, title, category, weight, max_score, is_published)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetGradeAssignmentFinalization :one
SELECT assignment_id, finalized_by, notes, finalized_at, updated_at
FROM grade_assignment_finalizations
WHERE assignment_id = $1;

-- name: UpsertGradeAssignmentFinalization :one
INSERT INTO grade_assignment_finalizations (assignment_id, finalized_by, notes, finalized_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (assignment_id) DO UPDATE
SET finalized_by = EXCLUDED.finalized_by,
    notes = EXCLUDED.notes,
    finalized_at = NOW(),
    updated_at = NOW()
RETURNING *;

-- name: DeleteGradeAssignmentFinalization :exec
DELETE FROM grade_assignment_finalizations WHERE assignment_id = $1;

-- name: UpdateGradeComponent :one
UPDATE grade_components
SET title = $2,
    category = $3,
    weight = $4,
    max_score = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateGradeComponentPublishState :one
UPDATE grade_components
SET is_published = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGradeComponent :exec
DELETE FROM grade_components WHERE id = $1;

-- name: ListGradebookSummary :many
WITH component_set AS (
  SELECT gc.id, gc.weight, gc.max_score
  FROM grade_components gc
  WHERE gc.assignment_id = $1
    AND (NOT sqlc.arg(published_only)::boolean OR gc.is_published = TRUE)
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
