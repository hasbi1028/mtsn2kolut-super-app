-- name: CreateCbtQuestionAsset :one
INSERT INTO cbt_question_assets (
  question_id, original_name, stored_name, mime_type, file_size, storage_path, purpose, uploaded_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetCbtQuestionAsset :one
SELECT id, question_id, original_name, stored_name, mime_type, file_size, storage_path, purpose, uploaded_by, created_at
FROM cbt_question_assets
WHERE id = $1;

-- name: ListCbtQuestionAssetsByQuestion :many
SELECT id, question_id, original_name, stored_name, mime_type, file_size, storage_path, purpose, uploaded_by, created_at
FROM cbt_question_assets
WHERE question_id = $1
ORDER BY created_at DESC;
