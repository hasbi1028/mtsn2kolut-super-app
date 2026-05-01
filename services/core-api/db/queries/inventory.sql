-- name: ListInventoryItems :many
SELECT id, kode, nama, kategori, lokasi, kondisi, satuan,
       jumlah_total, jumlah_baik, min_stock, catatan, created_at, updated_at
FROM inventory_items
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    kode     ILIKE '%' || sqlc.arg(search) || '%' OR
    nama     ILIKE '%' || sqlc.arg(search) || '%' OR
    lokasi   ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(kategori)::TEXT = '' OR kategori = sqlc.arg(kategori)
) AND (
    sqlc.arg(kondisi)::TEXT = '' OR kondisi = sqlc.arg(kondisi)
)
ORDER BY nama ASC;

-- name: GetInventoryItem :one
SELECT id, kode, nama, kategori, lokasi, kondisi, satuan,
       jumlah_total, jumlah_baik, min_stock, catatan, created_at, updated_at
FROM inventory_items
WHERE id = $1;

-- name: CreateInventoryItem :one
INSERT INTO inventory_items (
    kode, nama, kategori, lokasi, kondisi, satuan,
    jumlah_total, jumlah_baik, min_stock, catatan
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateInventoryItem :one
UPDATE inventory_items
SET kode         = $2,
    nama         = $3,
    kategori     = $4,
    lokasi       = $5,
    kondisi      = $6,
    satuan       = $7,
    jumlah_total = $8,
    jumlah_baik  = $9,
    min_stock    = $10,
    catatan      = $11,
    updated_at   = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteInventoryItem :exec
DELETE FROM inventory_items WHERE id = $1;

-- name: GetInventoryStats :one
SELECT
    (SELECT COUNT(*) FROM inventory_items) AS total_jenis,
    (SELECT COALESCE(SUM(jumlah_total), 0) FROM inventory_items) AS total_unit,
    (SELECT COALESCE(SUM(jumlah_baik), 0) FROM inventory_items) AS total_layak,
    (SELECT COUNT(*) FROM inventory_items WHERE jumlah_baik <= min_stock) AS perlu_restok,
    (SELECT COUNT(*) FROM inventory_items WHERE kondisi <> 'baik') AS perlu_perawatan;
