-- Library module foundation: book catalog + loan management

CREATE TYPE library_member_type AS ENUM ('student', 'employee');
CREATE TYPE loan_status_enum    AS ENUM ('active', 'returned');

CREATE TABLE library_books (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    kode            TEXT        NOT NULL UNIQUE,
    judul           TEXT        NOT NULL,
    pengarang       TEXT        NOT NULL DEFAULT '',
    isbn            TEXT        NOT NULL DEFAULT '',
    kategori        TEXT        NOT NULL DEFAULT 'umum',
    penerbit        TEXT        NOT NULL DEFAULT '',
    tahun_terbit    INT,
    total_eksemplar INT         NOT NULL DEFAULT 1,
    tersedia        INT         NOT NULL DEFAULT 1,
    lokasi_rak      TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_library_books_tersedia CHECK (tersedia >= 0 AND tersedia <= total_eksemplar)
);

CREATE TABLE library_loans (
    id              UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id         UUID                NOT NULL REFERENCES library_books(id) ON DELETE RESTRICT,
    member_type     library_member_type NOT NULL,
    student_id      UUID                REFERENCES students(id)  ON DELETE SET NULL,
    employee_id     UUID                REFERENCES employees(id) ON DELETE SET NULL,
    dipinjam_at     TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    jatuh_tempo     TIMESTAMPTZ         NOT NULL,
    dikembalikan_at TIMESTAMPTZ,
    denda_per_hari  INT                 NOT NULL DEFAULT 500,
    denda_total     INT                 NOT NULL DEFAULT 0,
    denda_lunas     BOOL                NOT NULL DEFAULT FALSE,
    status          loan_status_enum    NOT NULL DEFAULT 'active',
    catatan         TEXT                NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_library_loans_member CHECK (
        (member_type = 'student'  AND student_id  IS NOT NULL AND employee_id IS NULL) OR
        (member_type = 'employee' AND employee_id IS NOT NULL AND student_id  IS NULL)
    )
);

CREATE INDEX idx_library_books_kode     ON library_books(kode);
CREATE INDEX idx_library_books_kategori ON library_books(kategori);
CREATE INDEX idx_library_loans_book     ON library_loans(book_id);
CREATE INDEX idx_library_loans_student  ON library_loans(student_id)  WHERE student_id  IS NOT NULL;
CREATE INDEX idx_library_loans_employee ON library_loans(employee_id) WHERE employee_id IS NOT NULL;
CREATE INDEX idx_library_loans_status   ON library_loans(status, jatuh_tempo DESC);
