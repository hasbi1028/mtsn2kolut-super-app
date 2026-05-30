-- name: ListAssessmentExamPackageMaps :many
SELECT
  m.id,
  m.exam_id,
  m.class_id,
  c.code AS class_code,
  c.name AS class_name,
  CASE UPPER(NULLIF(btrim(c.level::text), ''))
    WHEN '7' THEN 7
    WHEN 'VII' THEN 7
    WHEN '8' THEN 8
    WHEN 'VIII' THEN 8
    WHEN '9' THEN 9
    WHEN 'IX' THEN 9
    ELSE 0
  END::int AS grade_level,
  m.subject_id,
  s.code AS subject_code,
  s.name AS subject_name,
  m.package_id,
  p.title AS package_title,
  COALESCE(p.duration_minutes, 0)::int AS duration_minutes,
  m.slot_label,
  m.notes,
  m.created_at,
  m.updated_at
FROM assessment_exam_package_maps m
JOIN school_classes c ON c.id = m.class_id
JOIN subjects s ON s.id = m.subject_id
JOIN cbt_packages p ON p.id = m.package_id
WHERE m.exam_id = sqlc.arg(exam_id)
ORDER BY grade_level, c.code, s.name, p.title;

-- name: UpsertAssessmentExamPackageMap :one
INSERT INTO assessment_exam_package_maps (exam_id, class_id, subject_id, package_id, slot_label, notes)
SELECT
  sqlc.arg(exam_id)::uuid,
  sqlc.arg(class_id)::uuid,
  sqlc.arg(subject_id)::uuid,
  p.id,
  sqlc.arg(slot_label)::text,
  sqlc.arg(notes)::text
FROM cbt_packages p
WHERE p.id = sqlc.arg(package_id)::uuid
  AND p.subject_id = sqlc.arg(subject_id)::uuid
  AND p.is_active = TRUE
  AND COALESCE(p.composition_log->>'archived', 'false') <> 'true'
ON CONFLICT (exam_id, class_id, subject_id) DO UPDATE
SET package_id = EXCLUDED.package_id,
    slot_label = EXCLUDED.slot_label,
    notes = EXCLUDED.notes,
    updated_at = NOW()
RETURNING *;

-- name: DeleteAssessmentExamPackageMap :execrows
DELETE FROM assessment_exam_package_maps
WHERE exam_id = sqlc.arg(exam_id)
  AND id = sqlc.arg(id);

-- name: ListAssessmentPackageOptions :many
SELECT
  p.id,
  p.event_id,
  p.subject_id,
  s.code AS subject_code,
  s.name AS subject_name,
  p.title,
  p.description,
  p.duration_minutes,
  p.is_active,
  COUNT(DISTINCT pq.question_id)::int AS question_count,
  COUNT(DISTINCT ses.id)::int AS session_count,
  p.locked_at,
  p.snapshot_version,
  p.created_at,
  p.updated_at
FROM cbt_packages p
JOIN subjects s ON s.id = p.subject_id
LEFT JOIN cbt_package_questions pq ON pq.package_id = p.id
LEFT JOIN cbt_exam_sessions ses ON ses.package_id = p.id
WHERE (sqlc.arg(subject_id)::uuid IS NULL OR p.subject_id = sqlc.arg(subject_id)::uuid)
  AND p.is_active = TRUE
  AND COALESCE(p.composition_log->>'archived', 'false') <> 'true'
GROUP BY p.id, s.code, s.name
ORDER BY s.name, p.created_at DESC;
