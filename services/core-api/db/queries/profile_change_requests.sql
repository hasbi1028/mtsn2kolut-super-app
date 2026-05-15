-- name: GetOwnedEmployeeOfficialProfile :one
SELECT
    e.id,
    e.nama,
    e.tanggal_lahir
FROM users u
JOIN employees e ON e.id = u.employee_id
WHERE u.id = $1
  AND u.deleted_at IS NULL;

-- name: GetOwnedStudentOfficialProfile :one
SELECT
    s.id,
    s.nama,
    s.tanggal_lahir,
    s.parent_name,
    s.phone,
    s.alamat
FROM users u
JOIN students s ON s.id = u.student_id
WHERE u.id = $1
  AND u.deleted_at IS NULL;

-- name: GetOwnedParentOfficialProfile :one
SELECT
    p.id,
    p.nama,
    p.phone,
    p.address,
    p.occupation,
    p.nik
FROM users u
JOIN parents p ON p.id = u.parent_id
WHERE u.id = $1
  AND u.deleted_at IS NULL;

-- name: GetParentOwnedChildOfficialProfile :one
SELECT
    s.id,
    s.nama,
    s.tanggal_lahir,
    s.parent_name,
    s.phone,
    s.alamat
FROM users u
JOIN parent_students ps ON ps.parent_id = u.parent_id
JOIN students s ON s.id = ps.student_id
WHERE u.id = sqlc.arg(requester_user_id)
  AND s.id = sqlc.arg(target_student_id)
  AND u.deleted_at IS NULL
  AND u.is_active = TRUE
  AND u.parent_id IS NOT NULL;

-- name: CreateProfileChangeRequest :one
INSERT INTO profile_change_requests (
    requester_user_id,
    profile_type,
    target_employee_id,
    target_student_id,
    target_parent_id,
    field_key,
    current_value,
    requested_value,
    reason
)
VALUES (
    sqlc.arg(requester_user_id),
    sqlc.arg(profile_type),
    sqlc.arg(target_employee_id),
    sqlc.arg(target_student_id),
    sqlc.arg(target_parent_id),
    sqlc.arg(field_key),
    sqlc.arg(current_value),
    sqlc.arg(requested_value),
    sqlc.arg(reason)
)
RETURNING *;

-- name: ListOwnProfileChangeRequests :many
SELECT
    pcr.*,
    req.username AS requester_username,
    COALESCE(NULLIF(btrim(req.display_name), ''), req_e.nama, req_s.nama, req_p.nama, e.nama, s.nama, p.nama, req.username)::text AS requester_display_name,
    reviewer.username AS reviewer_username,
    COALESCE(NULLIF(btrim(reviewer.display_name), ''), reviewer_e.nama, reviewer_s.nama, reviewer_p.nama, reviewer.username, '')::text AS reviewer_display_name,
    COALESCE(e.nama, s.nama, p.nama, '')::text AS profile_nama
FROM profile_change_requests pcr
JOIN users req ON req.id = pcr.requester_user_id
LEFT JOIN users reviewer ON reviewer.id = pcr.reviewer_user_id
LEFT JOIN employees req_e ON req_e.id = req.employee_id
LEFT JOIN students req_s ON req_s.id = req.student_id
LEFT JOIN parents req_p ON req_p.id = req.parent_id
LEFT JOIN employees reviewer_e ON reviewer_e.id = reviewer.employee_id
LEFT JOIN students reviewer_s ON reviewer_s.id = reviewer.student_id
LEFT JOIN parents reviewer_p ON reviewer_p.id = reviewer.parent_id
LEFT JOIN employees e ON e.id = pcr.target_employee_id
LEFT JOIN students s ON s.id = pcr.target_student_id
LEFT JOIN parents p ON p.id = pcr.target_parent_id
WHERE pcr.requester_user_id = $1
ORDER BY pcr.created_at DESC;

-- name: ListProfileChangeRequests :many
SELECT
    pcr.*,
    req.username AS requester_username,
    COALESCE(NULLIF(btrim(req.display_name), ''), req_e.nama, req_s.nama, req_p.nama, e.nama, s.nama, p.nama, req.username)::text AS requester_display_name,
    reviewer.username AS reviewer_username,
    COALESCE(NULLIF(btrim(reviewer.display_name), ''), reviewer_e.nama, reviewer_s.nama, reviewer_p.nama, reviewer.username, '')::text AS reviewer_display_name,
    COALESCE(e.nama, s.nama, p.nama, '')::text AS profile_nama
FROM profile_change_requests pcr
JOIN users req ON req.id = pcr.requester_user_id
LEFT JOIN users reviewer ON reviewer.id = pcr.reviewer_user_id
LEFT JOIN employees req_e ON req_e.id = req.employee_id
LEFT JOIN students req_s ON req_s.id = req.student_id
LEFT JOIN parents req_p ON req_p.id = req.parent_id
LEFT JOIN employees reviewer_e ON reviewer_e.id = reviewer.employee_id
LEFT JOIN students reviewer_s ON reviewer_s.id = reviewer.student_id
LEFT JOIN parents reviewer_p ON reviewer_p.id = reviewer.parent_id
LEFT JOIN employees e ON e.id = pcr.target_employee_id
LEFT JOIN students s ON s.id = pcr.target_student_id
LEFT JOIN parents p ON p.id = pcr.target_parent_id
WHERE (
    sqlc.arg(status_filter)::TEXT = ''
    OR pcr.status::TEXT = sqlc.arg(status_filter)::TEXT
)
AND (
    sqlc.arg(profile_type_filter)::TEXT = ''
    OR pcr.profile_type = sqlc.arg(profile_type_filter)::TEXT
)
AND (
    sqlc.arg(field_key_filter)::TEXT = ''
    OR pcr.field_key = sqlc.arg(field_key_filter)::TEXT
)
AND (
    sqlc.arg(search)::TEXT = ''
    OR req.username ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR COALESCE(NULLIF(btrim(req.display_name), ''), req_e.nama, req_s.nama, req_p.nama, e.nama, s.nama, p.nama, req.username)::TEXT ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR COALESCE(NULLIF(btrim(reviewer.display_name), ''), reviewer_e.nama, reviewer_s.nama, reviewer_p.nama, reviewer.username, '')::TEXT ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR COALESCE(e.nama, s.nama, p.nama, '')::TEXT ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR pcr.field_key ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR pcr.reason ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR pcr.review_note ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR pcr.current_value ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR pcr.requested_value ILIKE '%' || sqlc.arg(search)::TEXT || '%'
)
ORDER BY
    CASE WHEN pcr.status = 'pending' THEN 0 ELSE 1 END,
    pcr.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountProfileChangeRequests :one
SELECT COUNT(*)::INT
FROM profile_change_requests pcr
JOIN users req ON req.id = pcr.requester_user_id
LEFT JOIN users reviewer ON reviewer.id = pcr.reviewer_user_id
LEFT JOIN employees e ON e.id = pcr.target_employee_id
LEFT JOIN students s ON s.id = pcr.target_student_id
LEFT JOIN parents p ON p.id = pcr.target_parent_id
WHERE (
    sqlc.arg(status_filter)::TEXT = ''
    OR pcr.status::TEXT = sqlc.arg(status_filter)::TEXT
)
AND (
    sqlc.arg(profile_type_filter)::TEXT = ''
    OR pcr.profile_type = sqlc.arg(profile_type_filter)::TEXT
)
AND (
    sqlc.arg(field_key_filter)::TEXT = ''
    OR pcr.field_key = sqlc.arg(field_key_filter)::TEXT
)
AND (
    sqlc.arg(search)::TEXT = ''
    OR req.username ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR COALESCE(NULLIF(req.display_name, ''), e.nama, s.nama, p.nama, req.username)::TEXT ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR COALESCE(e.nama, s.nama, p.nama, '')::TEXT ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR pcr.field_key ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR pcr.reason ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR pcr.review_note ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR pcr.current_value ILIKE '%' || sqlc.arg(search)::TEXT || '%'
    OR pcr.requested_value ILIKE '%' || sqlc.arg(search)::TEXT || '%'
);

-- name: GetProfileChangeRequestForUpdate :one
SELECT *
FROM profile_change_requests
WHERE id = $1
FOR UPDATE;

-- name: CancelOwnProfileChangeRequest :one
UPDATE profile_change_requests
SET status = 'cancelled',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND requester_user_id = sqlc.arg(requester_user_id)
  AND status = 'pending'
RETURNING *;

-- name: ReviewProfileChangeRequest :one
UPDATE profile_change_requests
SET status = sqlc.arg(status),
    reviewer_user_id = sqlc.arg(reviewer_user_id),
    review_note = sqlc.arg(review_note),
    reviewed_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'pending'
RETURNING *;

-- name: UpdateEmployeeOfficialName :execrows
UPDATE employees
SET nama = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateEmployeeOfficialBirthdate :execrows
UPDATE employees
SET tanggal_lahir = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateStudentOfficialName :execrows
UPDATE students
SET nama = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateStudentOfficialBirthdate :execrows
UPDATE students
SET tanggal_lahir = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateStudentOfficialParentName :execrows
UPDATE students
SET parent_name = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateStudentOfficialPhone :execrows
UPDATE students
SET phone = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateStudentOfficialAddress :execrows
UPDATE students
SET alamat = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateParentOfficialName :execrows
UPDATE parents
SET nama = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateParentOfficialPhone :execrows
UPDATE parents
SET phone = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateParentOfficialAddress :execrows
UPDATE parents
SET address = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateParentOfficialOccupation :execrows
UPDATE parents
SET occupation = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateParentOfficialNik :execrows
UPDATE parents
SET nik = $2,
    updated_at = NOW()
WHERE id = $1;
