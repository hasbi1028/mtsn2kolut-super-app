-- =====================
-- Letter Classifications
-- =====================

-- name: ListLetterClassifications :many
SELECT code, name, description, is_active
FROM letter_classifications
WHERE is_active = TRUE
ORDER BY code ASC;

-- =====================
-- Incoming Letter Sequences
-- =====================

-- name: IssueIncomingLetterSequence :one
INSERT INTO incoming_letter_sequences (year, last_seq)
VALUES ($1, 1)
ON CONFLICT (year) DO UPDATE
    SET last_seq = incoming_letter_sequences.last_seq + 1
RETURNING last_seq;

-- =====================
-- Outgoing Letter Sequences
-- =====================

-- name: IssueOutgoingLetterSequence :one
INSERT INTO outgoing_letter_sequences (year, classification_code, last_seq)
VALUES ($1, $2, 1)
ON CONFLICT (year, classification_code) DO UPDATE
    SET last_seq = outgoing_letter_sequences.last_seq + 1
RETURNING last_seq;

-- =====================
-- Incoming Letters
-- =====================

-- name: ListIncomingLetters :many
SELECT
    il.id, il.nomor_surat, il.nomor_agenda, il.tanggal_surat, il.tanggal_terima,
    il.asal, il.perihal, il.sifat, il.file_path, il.catatan,
    il.status, il.received_by_employee_id, il.created_at, il.updated_at,
    COALESCE(e.nama, '') AS received_by_name,
    COUNT(ld.id)::int    AS disposisi_count
FROM incoming_letters il
LEFT JOIN employees e ON e.id = il.received_by_employee_id
LEFT JOIN letter_dispositions ld ON ld.incoming_letter_id = il.id
WHERE (sqlc.arg(search)::TEXT = '' OR
       il.perihal ILIKE '%' || sqlc.arg(search) || '%' OR
       il.asal    ILIKE '%' || sqlc.arg(search) || '%' OR
       il.nomor_surat ILIKE '%' || sqlc.arg(search) || '%' OR
       il.nomor_agenda ILIKE '%' || sqlc.arg(search) || '%'
)
AND (sqlc.arg(filter_status)::TEXT = '' OR il.status::TEXT = sqlc.arg(filter_status))
GROUP BY il.id, e.nama
ORDER BY il.tanggal_terima DESC, il.created_at DESC;

-- name: GetIncomingLetter :one
SELECT
    il.id, il.nomor_surat, il.nomor_agenda, il.tanggal_surat, il.tanggal_terima,
    il.asal, il.perihal, il.sifat, il.file_path, il.catatan,
    il.status, il.received_by_employee_id, il.created_at, il.updated_at,
    COALESCE(e.nama, '') AS received_by_name
FROM incoming_letters il
LEFT JOIN employees e ON e.id = il.received_by_employee_id
WHERE il.id = $1;

-- name: CreateIncomingLetter :one
INSERT INTO incoming_letters (
    nomor_surat, nomor_agenda, tanggal_surat, tanggal_terima,
    asal, perihal, sifat, file_path, catatan, received_by_employee_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, nomor_surat, nomor_agenda, tanggal_surat, tanggal_terima,
          asal, perihal, sifat, file_path, catatan, status,
          received_by_employee_id, created_at, updated_at;

-- name: UpdateIncomingLetter :one
UPDATE incoming_letters
SET nomor_surat             = $2,
    tanggal_surat           = $3,
    tanggal_terima          = $4,
    asal                    = $5,
    perihal                 = $6,
    sifat                   = $7,
    catatan                 = $8,
    updated_at              = NOW()
WHERE id = $1
RETURNING id, nomor_surat, nomor_agenda, tanggal_surat, tanggal_terima,
          asal, perihal, sifat, file_path, catatan, status,
          received_by_employee_id, created_at, updated_at;

-- name: UpdateIncomingLetterStatus :one
UPDATE incoming_letters
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, nomor_surat, nomor_agenda, tanggal_surat, tanggal_terima,
          asal, perihal, sifat, file_path, catatan, status,
          received_by_employee_id, created_at, updated_at;

-- name: DeleteIncomingLetter :exec
DELETE FROM incoming_letters WHERE id = $1;

-- =====================
-- Outgoing Letters
-- =====================

-- name: ListOutgoingLetters :many
SELECT
    ol.id, ol.nomor_surat, ol.classification_code, ol.tanggal_surat,
    ol.tujuan, ol.perihal, ol.sifat, ol.file_path, ol.catatan,
    ol.issued_by_employee_id, ol.created_at, ol.updated_at,
    COALESCE(e.nama, '') AS issued_by_name,
    COALESCE(lc.name, '') AS classification_name
FROM outgoing_letters ol
LEFT JOIN employees e ON e.id = ol.issued_by_employee_id
LEFT JOIN letter_classifications lc ON lc.code = ol.classification_code
WHERE (sqlc.arg(search)::TEXT = '' OR
       ol.perihal ILIKE '%' || sqlc.arg(search) || '%' OR
       ol.tujuan  ILIKE '%' || sqlc.arg(search) || '%' OR
       ol.nomor_surat ILIKE '%' || sqlc.arg(search) || '%'
)
ORDER BY ol.tanggal_surat DESC, ol.created_at DESC;

-- name: GetOutgoingLetter :one
SELECT
    ol.id, ol.nomor_surat, ol.classification_code, ol.tanggal_surat,
    ol.tujuan, ol.perihal, ol.sifat, ol.file_path, ol.catatan,
    ol.issued_by_employee_id, ol.created_at, ol.updated_at,
    COALESCE(e.nama, '')  AS issued_by_name,
    COALESCE(lc.name, '') AS classification_name
FROM outgoing_letters ol
LEFT JOIN employees e ON e.id = ol.issued_by_employee_id
LEFT JOIN letter_classifications lc ON lc.code = ol.classification_code
WHERE ol.id = $1;

-- name: CreateOutgoingLetter :one
INSERT INTO outgoing_letters (
    nomor_surat, classification_code, tanggal_surat,
    tujuan, perihal, sifat, file_path, catatan, issued_by_employee_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, nomor_surat, classification_code, tanggal_surat,
          tujuan, perihal, sifat, file_path, catatan,
          issued_by_employee_id, created_at, updated_at;

-- name: UpdateOutgoingLetter :one
UPDATE outgoing_letters
SET tanggal_surat = $2,
    tujuan        = $3,
    perihal       = $4,
    sifat         = $5,
    catatan       = $6,
    updated_at    = NOW()
WHERE id = $1
RETURNING id, nomor_surat, classification_code, tanggal_surat,
          tujuan, perihal, sifat, file_path, catatan,
          issued_by_employee_id, created_at, updated_at;

-- name: DeleteOutgoingLetter :exec
DELETE FROM outgoing_letters WHERE id = $1;

-- =====================
-- Dispositions
-- =====================

-- name: ListDispositions :many
SELECT
    ld.id, ld.incoming_letter_id, ld.assignee_employee_id,
    ld.instruksi, ld.catatan_tindak_lanjut, ld.status,
    ld.disposed_by_employee_id, ld.disposed_at, ld.completed_at,
    COALESCE(ae.nama, '') AS assignee_name,
    COALESCE(de.nama, '') AS disposed_by_name,
    il.nomor_agenda, il.perihal AS letter_perihal, il.asal AS letter_asal
FROM letter_dispositions ld
LEFT JOIN employees ae ON ae.id = ld.assignee_employee_id
LEFT JOIN employees de ON de.id = ld.disposed_by_employee_id
JOIN incoming_letters il ON il.id = ld.incoming_letter_id
WHERE (sqlc.arg(incoming_letter_id)::UUID IS NULL OR ld.incoming_letter_id = sqlc.arg(incoming_letter_id))
  AND (sqlc.arg(filter_status)::TEXT = '' OR ld.status::TEXT = sqlc.arg(filter_status))
ORDER BY ld.disposed_at DESC;

-- name: GetDisposition :one
SELECT
    ld.id, ld.incoming_letter_id, ld.assignee_employee_id,
    ld.instruksi, ld.catatan_tindak_lanjut, ld.status,
    ld.disposed_by_employee_id, ld.disposed_at, ld.completed_at,
    COALESCE(ae.nama, '') AS assignee_name,
    COALESCE(de.nama, '') AS disposed_by_name,
    il.nomor_agenda, il.perihal AS letter_perihal, il.asal AS letter_asal
FROM letter_dispositions ld
LEFT JOIN employees ae ON ae.id = ld.assignee_employee_id
LEFT JOIN employees de ON de.id = ld.disposed_by_employee_id
JOIN incoming_letters il ON il.id = ld.incoming_letter_id
WHERE ld.id = $1;

-- name: CreateDisposition :one
INSERT INTO letter_dispositions (
    incoming_letter_id, assignee_employee_id, instruksi,
    disposed_by_employee_id
) VALUES ($1, $2, $3, $4)
RETURNING id, incoming_letter_id, assignee_employee_id, instruksi,
          catatan_tindak_lanjut, status, disposed_by_employee_id,
          disposed_at, completed_at;

-- name: UpdateDisposition :one
UPDATE letter_dispositions
SET instruksi               = $2,
    catatan_tindak_lanjut   = $3,
    status                  = $4,
    completed_at            = CASE WHEN $4::disposition_status = 'selesai' THEN NOW() ELSE completed_at END
WHERE id = $1
RETURNING id, incoming_letter_id, assignee_employee_id, instruksi,
          catatan_tindak_lanjut, status, disposed_by_employee_id,
          disposed_at, completed_at;

-- name: DeleteDisposition :exec
DELETE FROM letter_dispositions WHERE id = $1;

-- name: CountDispositionsForLetter :one
SELECT COUNT(*) FROM letter_dispositions WHERE incoming_letter_id = $1;
