-- name: GetArchiveStats :one
SELECT
    (SELECT COUNT(*) FROM archive_categories WHERE is_active)::BIGINT AS active_categories,
    (SELECT COUNT(*) FROM archive_documents)::BIGINT AS total_documents,
    (SELECT COUNT(*) FROM archive_documents WHERE status = 'active')::BIGINT AS active_documents,
    (SELECT COUNT(*) FROM archive_documents WHERE status = 'borrowed')::BIGINT AS borrowed_documents,
    (SELECT COUNT(*) FROM archive_documents WHERE status = 'disposed')::BIGINT AS disposed_documents,
    (SELECT COUNT(*) FROM archive_documents WHERE EXTRACT(YEAR FROM received_date) = EXTRACT(YEAR FROM NOW()))::BIGINT AS documents_this_year,
    (SELECT COALESCE(SUM(file_size), 0) FROM archive_documents)::BIGINT AS total_file_size;

-- name: ListArchiveCategories :many
SELECT
    ac.id,
    ac.code,
    ac.name,
    COALESCE(ac.classification_code, '')::TEXT AS classification_code,
    COALESCE(lc.name, '')::TEXT AS classification_name,
    ac.description,
    ac.retention_years,
    ac.is_active,
    ac.created_at,
    ac.updated_at,
    COALESCE(doc.document_count, 0)::BIGINT AS document_count
FROM archive_categories ac
LEFT JOIN letter_classifications lc ON lc.code = ac.classification_code
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS document_count
    FROM archive_documents ad
    WHERE ad.category_id = ac.id
) doc ON TRUE
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    ac.code ILIKE '%' || sqlc.arg(search) || '%' OR
    ac.name ILIKE '%' || sqlc.arg(search) || '%' OR
    ac.description ILIKE '%' || sqlc.arg(search) || '%' OR
    ac.classification_code ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(active_only)::BOOLEAN = FALSE OR ac.is_active = TRUE
)
ORDER BY ac.is_active DESC, ac.code;

-- name: GetArchiveCategory :one
SELECT *
FROM archive_categories
WHERE id = $1;

-- name: CreateArchiveCategory :one
INSERT INTO archive_categories (
    code, name, classification_code, description, retention_years, is_active
) VALUES (
    sqlc.arg(code),
    sqlc.arg(name),
    NULLIF(sqlc.arg(classification_code)::TEXT, ''),
    sqlc.arg(description),
    sqlc.arg(retention_years),
    sqlc.arg(is_active)
)
RETURNING *;

-- name: UpdateArchiveCategory :one
UPDATE archive_categories
SET code = sqlc.arg(code),
    name = sqlc.arg(name),
    classification_code = NULLIF(sqlc.arg(classification_code)::TEXT, ''),
    description = sqlc.arg(description),
    retention_years = sqlc.arg(retention_years),
    is_active = sqlc.arg(is_active),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteArchiveCategory :exec
DELETE FROM archive_categories WHERE id = $1;

-- name: ListArchiveDocuments :many
SELECT
    ad.id,
    ad.category_id,
    ac.code AS category_code,
    ac.name AS category_name,
    COALESCE(ac.classification_code, '')::TEXT AS classification_code,
    ad.title,
    ad.archive_number,
    ad.document_date,
    ad.received_date,
    ad.summary,
    ad.tags,
    ad.status,
    ad.storage_location,
    ad.retention_until,
    ad.original_name,
    ad.stored_name,
    ad.mime_type,
    ad.file_size,
    ad.checksum_sha256,
    ad.uploaded_by_user_id,
    COALESCE(u.username, '')::TEXT AS uploaded_by_username,
    ad.created_at,
    ad.updated_at
FROM archive_documents ad
JOIN archive_categories ac ON ac.id = ad.category_id
LEFT JOIN users u ON u.id = ad.uploaded_by_user_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    ad.title ILIKE '%' || sqlc.arg(search) || '%' OR
    ad.archive_number ILIKE '%' || sqlc.arg(search) || '%' OR
    ad.summary ILIKE '%' || sqlc.arg(search) || '%' OR
    ad.tags ILIKE '%' || sqlc.arg(search) || '%' OR
    ad.original_name ILIKE '%' || sqlc.arg(search) || '%' OR
    ac.code ILIKE '%' || sqlc.arg(search) || '%' OR
    ac.name ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(category_id)::UUID IS NULL OR ad.category_id = sqlc.arg(category_id)::UUID
) AND (
    sqlc.arg(status)::TEXT = '' OR ad.status = sqlc.arg(status)
) AND (
    sqlc.arg(classification_code)::TEXT = '' OR ac.classification_code = sqlc.arg(classification_code)
)
ORDER BY ad.received_date DESC, ad.created_at DESC;

-- name: GetArchiveDocument :one
SELECT *
FROM archive_documents
WHERE id = $1;

-- name: GetArchiveDocumentDetail :one
SELECT
    ad.id,
    ad.category_id,
    ac.code AS category_code,
    ac.name AS category_name,
    COALESCE(ac.classification_code, '')::TEXT AS classification_code,
    ad.title,
    ad.archive_number,
    ad.document_date,
    ad.received_date,
    ad.summary,
    ad.tags,
    ad.status,
    ad.storage_location,
    ad.retention_until,
    ad.original_name,
    ad.stored_name,
    ad.mime_type,
    ad.file_size,
    ad.checksum_sha256,
    ad.uploaded_by_user_id,
    COALESCE(u.username, '')::TEXT AS uploaded_by_username,
    ad.created_at,
    ad.updated_at
FROM archive_documents ad
JOIN archive_categories ac ON ac.id = ad.category_id
LEFT JOIN users u ON u.id = ad.uploaded_by_user_id
WHERE ad.id = $1;

-- name: CreateArchiveDocument :one
INSERT INTO archive_documents (
    category_id,
    title,
    archive_number,
    document_date,
    received_date,
    summary,
    tags,
    status,
    storage_location,
    retention_until,
    original_name,
    stored_name,
    file_path,
    mime_type,
    file_size,
    checksum_sha256,
    uploaded_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: UpdateArchiveDocument :one
UPDATE archive_documents
SET category_id = $2,
    title = $3,
    archive_number = $4,
    document_date = $5,
    received_date = $6,
    summary = $7,
    tags = $8,
    status = $9,
    storage_location = $10,
    retention_until = $11,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteArchiveDocument :exec
DELETE FROM archive_documents WHERE id = $1;
