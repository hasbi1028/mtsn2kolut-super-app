-- =====================
-- Books
-- =====================

-- name: ListBooks :many
SELECT id, kode, judul, pengarang, isbn, kategori, penerbit, tahun_terbit,
       total_eksemplar, tersedia, lokasi_rak, created_at, updated_at
FROM library_books
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    judul     ILIKE '%' || sqlc.arg(search) || '%' OR
    pengarang ILIKE '%' || sqlc.arg(search) || '%' OR
    kode      ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(kategori)::TEXT = '' OR kategori = sqlc.arg(kategori)
)
ORDER BY judul ASC;

-- name: GetBook :one
SELECT id, kode, judul, pengarang, isbn, kategori, penerbit, tahun_terbit,
       total_eksemplar, tersedia, lokasi_rak, created_at, updated_at
FROM library_books
WHERE id = $1;

-- name: CreateBook :one
INSERT INTO library_books (kode, judul, pengarang, isbn, kategori, penerbit, tahun_terbit, total_eksemplar, tersedia, lokasi_rak)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, $9)
RETURNING *;

-- name: UpdateBook :one
UPDATE library_books
SET kode            = $2,
    judul           = $3,
    pengarang       = $4,
    isbn            = $5,
    kategori        = $6,
    penerbit        = $7,
    tahun_terbit    = $8,
    total_eksemplar = $9,
    lokasi_rak      = $10,
    updated_at      = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteBook :exec
DELETE FROM library_books WHERE id = $1;

-- name: DecrementTersedia :exec
UPDATE library_books SET tersedia = tersedia - 1, updated_at = NOW() WHERE id = $1;

-- name: IncrementTersedia :exec
UPDATE library_books SET tersedia = tersedia + 1, updated_at = NOW() WHERE id = $1;

-- =====================
-- Loans
-- =====================

-- name: ListLoans :many
SELECT
    l.id, l.book_id, l.member_type, l.student_id, l.employee_id,
    l.dipinjam_at, l.jatuh_tempo, l.dikembalikan_at,
    l.denda_per_hari, l.denda_total, l.denda_lunas,
    l.status, l.catatan, l.created_at, l.updated_at,
    b.kode      AS book_kode,
    b.judul     AS book_judul,
    COALESCE(s.nama, e.nama, '')  AS member_nama,
    COALESCE(s.nis,  e.nip, '')  AS member_nip_nis,
    (l.status = 'active' AND l.jatuh_tempo < NOW()) AS is_overdue
FROM library_loans l
JOIN  library_books b ON b.id = l.book_id
LEFT JOIN students  s ON s.id = l.student_id
LEFT JOIN employees e ON e.id = l.employee_id
WHERE (sqlc.arg(filter_status)::TEXT = '' OR l.status::TEXT = sqlc.arg(filter_status))
ORDER BY l.dipinjam_at DESC;

-- name: GetLoan :one
SELECT
    l.id, l.book_id, l.member_type, l.student_id, l.employee_id,
    l.dipinjam_at, l.jatuh_tempo, l.dikembalikan_at,
    l.denda_per_hari, l.denda_total, l.denda_lunas,
    l.status, l.catatan, l.created_at, l.updated_at,
    b.kode  AS book_kode,
    b.judul AS book_judul,
    COALESCE(s.nama, e.nama, '') AS member_nama,
    COALESCE(s.nis,  e.nip, '') AS member_nip_nis,
    (l.status = 'active' AND l.jatuh_tempo < NOW()) AS is_overdue
FROM library_loans l
JOIN  library_books b ON b.id = l.book_id
LEFT JOIN students  s ON s.id = l.student_id
LEFT JOIN employees e ON e.id = l.employee_id
WHERE l.id = $1;

-- name: CreateLoan :one
INSERT INTO library_loans (book_id, member_type, student_id, employee_id, jatuh_tempo, denda_per_hari)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateLoanReturn :one
UPDATE library_loans
SET dikembalikan_at = NOW(),
    denda_total     = $2,
    status          = 'returned',
    updated_at      = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkLoanDendaLunas :one
UPDATE library_loans
SET denda_lunas = TRUE, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountActiveLoansForMember :one
SELECT COUNT(*) FROM library_loans
WHERE status = 'active'
  AND (
    (member_type = 'student'  AND student_id  = sqlc.arg(member_id)) OR
    (member_type = 'employee' AND employee_id = sqlc.arg(member_id))
  );

-- name: GetLibraryStats :one
SELECT
    (SELECT COUNT(*)                    FROM library_books)                                                    AS total_judul,
    (SELECT COALESCE(SUM(total_eksemplar), 0) FROM library_books)                                             AS total_eksemplar,
    (SELECT COALESCE(SUM(tersedia), 0)  FROM library_books)                                                   AS total_tersedia,
    (SELECT COUNT(*)                    FROM library_loans WHERE status = 'active')                            AS sedang_dipinjam,
    (SELECT COUNT(*)                    FROM library_loans WHERE status = 'active' AND jatuh_tempo < NOW())    AS terlambat,
    (SELECT COUNT(*)                    FROM library_loans WHERE status = 'returned' AND NOT denda_lunas AND denda_total > 0) AS denda_belum_lunas;
