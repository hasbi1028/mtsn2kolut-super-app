-- name: ListCertificateTemplates :many
SELECT
    id,
    code,
    name,
    description,
    default_purpose,
    body_template,
    is_active,
    created_at,
    updated_at
FROM certificate_templates
WHERE (@active_only::bool = FALSE OR is_active = TRUE)
ORDER BY name;

-- name: GetCertificateTemplateByID :one
SELECT
    id,
    code,
    name,
    description,
    default_purpose,
    body_template,
    is_active,
    created_at,
    updated_at
FROM certificate_templates
WHERE id = $1;

-- name: ListStudentCertificateOptions :many
SELECT
    s.id,
    s.nis,
    s.nisn,
    s.nama,
    s.gender,
    s.status,
    s.class_id,
    COALESCE(sc.name, '')::text AS class_name,
    COALESCE(sc.level, 0)::int AS class_level,
    s.parent_name,
    s.parent_phone,
    s.nik,
    s.tempat_lahir,
    s.tanggal_lahir,
    s.alamat,
    s.agama,
    s.phone
FROM students s
LEFT JOIN school_classes sc ON sc.id = s.class_id
WHERE (@search::text = ''
       OR s.nama ILIKE '%' || @search || '%'
       OR s.nis ILIKE '%' || @search || '%'
       OR s.nisn ILIKE '%' || @search || '%')
  AND (@status::text = '' OR s.status::text = @status)
ORDER BY s.nama
LIMIT 80;

-- name: GetCertificateStudentSnapshot :one
SELECT
    s.id,
    s.nis,
    s.nisn,
    s.nama,
    s.gender,
    s.status,
    s.class_id,
    COALESCE(sc.name, '')::text AS class_name,
    COALESCE(sc.level, 0)::int AS class_level,
    s.parent_name,
    s.parent_phone,
    s.nik,
    s.tempat_lahir,
    s.tanggal_lahir,
    s.alamat,
    s.agama,
    s.phone
FROM students s
LEFT JOIN school_classes sc ON sc.id = s.class_id
WHERE s.id = $1;

-- name: ListStudentCertificates :many
SELECT
    c.id,
    c.template_id,
    t.code AS template_code,
    t.name AS template_name,
    c.student_id,
    s.nis AS student_nis,
    s.nisn AS student_nisn,
    s.nama AS student_name,
    COALESCE(sc.name, '')::text AS class_name,
    c.outgoing_letter_id,
    o.nomor_surat,
    o.classification_code,
    c.tanggal_surat,
    c.purpose,
    c.recipient,
    c.remarks,
    c.status,
    COALESCE(u.username, '')::text AS created_by_username,
    c.created_at,
    c.updated_at
FROM student_certificates c
JOIN certificate_templates t ON t.id = c.template_id
JOIN students s ON s.id = c.student_id
LEFT JOIN school_classes sc ON sc.id = s.class_id
JOIN outgoing_letters o ON o.id = c.outgoing_letter_id
LEFT JOIN users u ON u.id = c.created_by_user_id
WHERE (@search::text = ''
       OR s.nama ILIKE '%' || @search || '%'
       OR s.nis ILIKE '%' || @search || '%'
       OR s.nisn ILIKE '%' || @search || '%'
       OR o.nomor_surat ILIKE '%' || @search || '%'
       OR c.purpose ILIKE '%' || @search || '%')
  AND (@status::text = '' OR c.status = @status)
  AND (@template_code::text = '' OR t.code = @template_code)
ORDER BY c.tanggal_surat DESC, c.created_at DESC;

-- name: GetStudentCertificate :one
SELECT
    c.id,
    c.template_id,
    t.code AS template_code,
    t.name AS template_name,
    t.description AS template_description,
    t.default_purpose AS template_default_purpose,
    t.body_template AS template_body,
    c.student_id,
    s.nis AS student_nis,
    s.nisn AS student_nisn,
    s.nama AS student_name,
    s.gender AS student_gender,
    s.status AS student_status,
    s.class_id AS student_class_id,
    COALESCE(sc.name, '')::text AS class_name,
    COALESCE(sc.level, 0)::int AS class_level,
    s.parent_name,
    s.parent_phone,
    s.nik,
    s.tempat_lahir,
    s.tanggal_lahir,
    s.alamat,
    s.agama,
    s.phone,
    c.outgoing_letter_id,
    o.nomor_surat,
    o.classification_code,
    o.tujuan,
    o.perihal,
    c.tanggal_surat,
    c.purpose,
    c.recipient,
    c.remarks,
    c.snapshot_data,
    c.status,
    COALESCE(u.username, '')::text AS created_by_username,
    c.created_at,
    c.updated_at
FROM student_certificates c
JOIN certificate_templates t ON t.id = c.template_id
JOIN students s ON s.id = c.student_id
LEFT JOIN school_classes sc ON sc.id = s.class_id
JOIN outgoing_letters o ON o.id = c.outgoing_letter_id
LEFT JOIN users u ON u.id = c.created_by_user_id
WHERE c.id = $1;

-- name: CreateStudentCertificate :one
INSERT INTO student_certificates (
    template_id,
    student_id,
    outgoing_letter_id,
    tanggal_surat,
    purpose,
    recipient,
    remarks,
    snapshot_data,
    created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING
    id,
    template_id,
    student_id,
    outgoing_letter_id,
    tanggal_surat,
    purpose,
    recipient,
    remarks,
    snapshot_data,
    status,
    created_by_user_id,
    created_at,
    updated_at;

-- name: CancelStudentCertificate :one
UPDATE student_certificates
SET status = 'canceled',
    remarks = CASE
        WHEN @remarks::text = '' THEN remarks
        WHEN remarks = '' THEN @remarks
        ELSE remarks || E'\n' || @remarks
    END,
    updated_at = NOW()
WHERE id = @id
RETURNING
    id,
    template_id,
    student_id,
    outgoing_letter_id,
    tanggal_surat,
    purpose,
    recipient,
    remarks,
    snapshot_data,
    status,
    created_by_user_id,
    created_at,
    updated_at;
