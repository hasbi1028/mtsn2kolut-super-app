-- name: ListWebsiteContents :many
SELECT id, kind, title, slug, excerpt, content_html, cover_image_url, is_featured, meta_title, meta_description, status, published_at, created_by, updated_by, created_at, updated_at
FROM website_contents
WHERE (
    sqlc.arg(kind_filter)::TEXT = '' OR kind::TEXT = sqlc.arg(kind_filter)
) AND (
    sqlc.arg(status_filter)::TEXT = '' OR status::TEXT = sqlc.arg(status_filter)
) AND (
    sqlc.arg(search_query)::TEXT = '' OR
    title ILIKE '%' || sqlc.arg(search_query) || '%' OR
    slug ILIKE '%' || sqlc.arg(search_query) || '%' OR
    excerpt ILIKE '%' || sqlc.arg(search_query) || '%'
)
ORDER BY
    CASE status::TEXT WHEN 'published' THEN 0 ELSE 1 END,
    COALESCE(published_at, created_at) DESC,
    created_at DESC;

-- name: GetWebsiteContent :one
SELECT id, kind, title, slug, excerpt, content_html, cover_image_url, is_featured, meta_title, meta_description, status, published_at, created_by, updated_by, created_at, updated_at
FROM website_contents
WHERE id = $1;

-- name: CreateWebsiteContent :one
INSERT INTO website_contents (
    kind, title, slug, excerpt, content_html, cover_image_url, is_featured, meta_title, meta_description, status, published_at, created_by, updated_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: UpdateWebsiteContent :one
UPDATE website_contents
SET kind            = $2,
    title           = $3,
    slug            = $4,
    excerpt         = $5,
    content_html    = $6,
    cover_image_url = $7,
    is_featured     = $8,
    meta_title      = $9,
    meta_description = $10,
    status          = $11,
    published_at    = $12,
    updated_by      = $13,
    updated_at      = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteWebsiteContent :exec
DELETE FROM website_contents
WHERE id = $1;

-- name: ListPublishedWebsiteContents :many
SELECT id, kind, title, slug, excerpt, content_html, cover_image_url, is_featured, meta_title, meta_description, status, published_at, created_by, updated_by, created_at, updated_at
FROM website_contents
WHERE kind::TEXT = sqlc.arg(kind_filter)
  AND status = 'published'
ORDER BY is_featured DESC, COALESCE(published_at, created_at) DESC, created_at DESC
LIMIT sqlc.arg(limit_count);

-- name: GetPublishedWebsiteContentBySlug :one
SELECT id, kind, title, slug, excerpt, content_html, cover_image_url, is_featured, meta_title, meta_description, status, published_at, created_by, updated_by, created_at, updated_at
FROM website_contents
WHERE kind::TEXT = sqlc.arg(kind_filter)
  AND slug = sqlc.arg(slug_value)
  AND status = 'published';

-- name: ListFeaturedWebsiteContents :many
SELECT id, kind, title, slug, excerpt, content_html, cover_image_url, is_featured, meta_title, meta_description, status, published_at, created_by, updated_by, created_at, updated_at
FROM website_contents
WHERE kind::TEXT = sqlc.arg(kind_filter)
  AND status = 'published'
  AND is_featured = TRUE
ORDER BY COALESCE(published_at, created_at) DESC
LIMIT sqlc.arg(limit_count);
