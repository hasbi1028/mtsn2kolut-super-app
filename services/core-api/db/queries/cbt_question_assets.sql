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

-- name: BindCbtQuestionAssetsToQuestion :exec
UPDATE cbt_question_assets
SET question_id = sqlc.arg(question_id)::uuid
WHERE id = ANY(sqlc.arg(asset_ids)::uuid[])
  AND (question_id IS NULL OR question_id = sqlc.arg(question_id)::uuid);
