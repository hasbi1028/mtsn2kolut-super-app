-- name: CountKesiswaanStudents :one
SELECT COUNT(*)::BIGINT
FROM students s
LEFT JOIN school_classes c ON c.id = s.class_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    s.nis ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nisn ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nama ILIKE '%' || sqlc.arg(search) || '%' OR
    s.nik ILIKE '%' || sqlc.arg(search) || '%' OR
    c.name ILIKE '%' || sqlc.arg(search) || '%' OR
    c.code ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(status)::TEXT = '' OR s.status::TEXT = sqlc.arg(status)
) AND (
    sqlc.arg(class_id)::UUID IS NULL OR s.class_id = sqlc.arg(class_id)::UUID
) AND (
    sqlc.arg(teacher_employee_id)::UUID IS NULL OR EXISTS (
        SELECT 1
        FROM class_subject_assignments csa
        WHERE csa.class_id = s.class_id
          AND csa.teacher_employee_id = sqlc.arg(teacher_employee_id)::UUID
    )
);
