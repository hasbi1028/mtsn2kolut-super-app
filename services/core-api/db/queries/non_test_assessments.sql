-- name: ListNonTestAssessments :many
SELECT
  a.id,
  a.subject_id,
  s.name AS subject_name,
  s.code AS subject_code,
  a.class_id,
  COALESCE(sc.name, '') AS class_name,
  COALESCE(sc.level, '') AS class_level,
  a.assessment_type,
  a.title,
  a.description,
  a.instruction_html,
  a.rubric_html,
  a.evidence_requirements,
  a.mode,
  a.scoring_scale,
  a.max_score,
  a.weight,
  a.due_at,
  a.status,
  a.created_by_username,
  a.assessor_username,
  a.checklist,
  a.created_at,
  a.updated_at,
  COALESCE(submission_stats.total_submissions, 0)::int AS total_submissions,
  COALESCE(submission_stats.reviewed_submissions, 0)::int AS reviewed_submissions
FROM non_test_assessments a
JOIN subjects s ON s.id = a.subject_id
LEFT JOIN school_classes sc ON sc.id = a.class_id
LEFT JOIN LATERAL (
  SELECT
    COUNT(*)::int AS total_submissions,
    COUNT(*) FILTER (WHERE nas.status = 'reviewed')::int AS reviewed_submissions
  FROM non_test_assessment_submissions nas
  WHERE nas.assessment_id = a.id
) submission_stats ON TRUE
WHERE (sqlc.arg(subject_id)::uuid IS NULL OR a.subject_id = sqlc.arg(subject_id)::uuid)
  AND (sqlc.arg(class_id)::uuid IS NULL OR a.class_id = sqlc.arg(class_id)::uuid)
  AND (sqlc.arg(status_filter)::text = '' OR a.status = sqlc.arg(status_filter)::text)
  AND (sqlc.arg(assessment_type_filter)::text = '' OR a.assessment_type = sqlc.arg(assessment_type_filter)::text)
  AND (
    sqlc.arg(search_query)::text = ''
    OR a.title ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR a.description ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR a.evidence_requirements ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR s.name ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR COALESCE(sc.name, '') ILIKE '%' || sqlc.arg(search_query)::text || '%'
  )
ORDER BY a.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountNonTestAssessments :one
SELECT COUNT(*)::bigint
FROM non_test_assessments a
JOIN subjects s ON s.id = a.subject_id
LEFT JOIN school_classes sc ON sc.id = a.class_id
WHERE (sqlc.arg(subject_id)::uuid IS NULL OR a.subject_id = sqlc.arg(subject_id)::uuid)
  AND (sqlc.arg(class_id)::uuid IS NULL OR a.class_id = sqlc.arg(class_id)::uuid)
  AND (sqlc.arg(status_filter)::text = '' OR a.status = sqlc.arg(status_filter)::text)
  AND (sqlc.arg(assessment_type_filter)::text = '' OR a.assessment_type = sqlc.arg(assessment_type_filter)::text)
  AND (
    sqlc.arg(search_query)::text = ''
    OR a.title ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR a.description ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR a.evidence_requirements ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR s.name ILIKE '%' || sqlc.arg(search_query)::text || '%'
    OR COALESCE(sc.name, '') ILIKE '%' || sqlc.arg(search_query)::text || '%'
  );

-- name: GetNonTestAssessment :one
SELECT
  a.id,
  a.subject_id,
  s.name AS subject_name,
  s.code AS subject_code,
  a.class_id,
  COALESCE(sc.name, '') AS class_name,
  COALESCE(sc.level, '') AS class_level,
  a.assessment_type,
  a.title,
  a.description,
  a.instruction_html,
  a.rubric_html,
  a.evidence_requirements,
  a.mode,
  a.scoring_scale,
  a.max_score,
  a.weight,
  a.due_at,
  a.status,
  a.created_by_username,
  a.assessor_username,
  a.checklist,
  a.created_at,
  a.updated_at,
  COALESCE(submission_stats.total_submissions, 0)::int AS total_submissions,
  COALESCE(submission_stats.reviewed_submissions, 0)::int AS reviewed_submissions
FROM non_test_assessments a
JOIN subjects s ON s.id = a.subject_id
LEFT JOIN school_classes sc ON sc.id = a.class_id
LEFT JOIN LATERAL (
  SELECT
    COUNT(*)::int AS total_submissions,
    COUNT(*) FILTER (WHERE nas.status = 'reviewed')::int AS reviewed_submissions
  FROM non_test_assessment_submissions nas
  WHERE nas.assessment_id = a.id
) submission_stats ON TRUE
WHERE a.id = $1;

-- name: CreateNonTestAssessment :one
INSERT INTO non_test_assessments (
  subject_id,
  class_id,
  assessment_type,
  title,
  description,
  instruction_html,
  rubric_html,
  evidence_requirements,
  mode,
  scoring_scale,
  max_score,
  weight,
  due_at,
  status,
  created_by_username,
  assessor_username,
  checklist
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING *;

-- name: UpdateNonTestAssessment :one
UPDATE non_test_assessments
SET
  subject_id = $2,
  class_id = $3,
  assessment_type = $4,
  title = $5,
  description = $6,
  instruction_html = $7,
  rubric_html = $8,
  evidence_requirements = $9,
  mode = $10,
  scoring_scale = $11,
  max_score = $12,
  weight = $13,
  due_at = $14,
  status = $15,
  created_by_username = $16,
  assessor_username = $17,
  checklist = $18,
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteNonTestAssessment :exec
DELETE FROM non_test_assessments WHERE id = $1;

-- name: ListNonTestSubmissions :many
SELECT
  nas.id,
  nas.assessment_id,
  nas.student_id,
  st.nis,
  st.nisn,
  st.nama AS student_name,
  st.class_id,
  COALESCE(sc.name, '') AS class_name,
  COALESCE(sc.level, '') AS class_level,
  nas.status,
  nas.evidence_url,
  nas.evidence_note,
  nas.score,
  nas.feedback,
  nas.submitted_at,
  nas.graded_at,
  nas.graded_by_username,
  nas.created_at,
  nas.updated_at
FROM non_test_assessment_submissions nas
JOIN students st ON st.id = nas.student_id
LEFT JOIN school_classes sc ON sc.id = st.class_id
WHERE nas.assessment_id = $1
ORDER BY sc.level NULLS LAST, sc.name NULLS LAST, st.nama;

-- name: UpsertNonTestSubmission :one
INSERT INTO non_test_assessment_submissions (
  assessment_id,
  student_id,
  status,
  evidence_url,
  evidence_note,
  score,
  feedback,
  submitted_at,
  graded_at,
  graded_by_username
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (assessment_id, student_id) DO UPDATE
SET
  status = EXCLUDED.status,
  evidence_url = EXCLUDED.evidence_url,
  evidence_note = EXCLUDED.evidence_note,
  score = EXCLUDED.score,
  feedback = EXCLUDED.feedback,
  submitted_at = EXCLUDED.submitted_at,
  graded_at = EXCLUDED.graded_at,
  graded_by_username = EXCLUDED.graded_by_username,
  updated_at = NOW()
RETURNING *;
