-- Sprint 17: Tata Usaha — Persuratan Core
-- letter_sifat, letter_status, disposition_status enums + 6 tables

CREATE TYPE letter_sifat AS ENUM ('biasa', 'penting', 'segera', 'rahasia');
CREATE TYPE letter_status AS ENUM ('baru', 'didisposisi', 'selesai', 'arsip');
CREATE TYPE disposition_status AS ENUM ('terkirim', 'dibaca', 'ditindaklanjuti', 'selesai');

-- Kode klasifikasi surat (Kemenag standard)
CREATE TABLE letter_classifications (
    code        TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE
);

-- Sequence nomor agenda surat masuk (reset per tahun)
CREATE TABLE incoming_letter_sequences (
    year     INT PRIMARY KEY,
    last_seq INT NOT NULL DEFAULT 0
);

-- Sequence nomor surat keluar per (tahun, kode klasifikasi)
CREATE TABLE outgoing_letter_sequences (
    year                INT  NOT NULL,
    classification_code TEXT NOT NULL REFERENCES letter_classifications(code),
    last_seq            INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (year, classification_code)
);

-- Surat masuk
CREATE TABLE incoming_letters (
    id                      UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    nomor_surat             TEXT         NOT NULL,
    nomor_agenda            TEXT         NOT NULL,
    tanggal_surat           DATE         NOT NULL,
    tanggal_terima          DATE         NOT NULL,
    asal                    TEXT         NOT NULL,
    perihal                 TEXT         NOT NULL,
    sifat                   letter_sifat NOT NULL DEFAULT 'biasa',
    file_path               TEXT         NOT NULL DEFAULT '',
    catatan                 TEXT         NOT NULL DEFAULT '',
    status                  letter_status NOT NULL DEFAULT 'baru',
    received_by_employee_id UUID         REFERENCES employees(id),
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_incoming_agenda UNIQUE (nomor_agenda)
);

-- Surat keluar
CREATE TABLE outgoing_letters (
    id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    nomor_surat           TEXT         NOT NULL,
    classification_code   TEXT         NOT NULL REFERENCES letter_classifications(code),
    tanggal_surat         DATE         NOT NULL,
    tujuan                TEXT         NOT NULL,
    perihal               TEXT         NOT NULL,
    sifat                 letter_sifat NOT NULL DEFAULT 'biasa',
    file_path             TEXT         NOT NULL DEFAULT '',
    catatan               TEXT         NOT NULL DEFAULT '',
    issued_by_employee_id UUID         REFERENCES employees(id),
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_outgoing_nomor UNIQUE (nomor_surat)
);

-- Disposisi surat masuk
CREATE TABLE letter_dispositions (
    id                       UUID               PRIMARY KEY DEFAULT gen_random_uuid(),
    incoming_letter_id       UUID               NOT NULL REFERENCES incoming_letters(id) ON DELETE CASCADE,
    assignee_employee_id     UUID               NOT NULL REFERENCES employees(id),
    instruksi                TEXT               NOT NULL DEFAULT '',
    catatan_tindak_lanjut    TEXT               NOT NULL DEFAULT '',
    status                   disposition_status NOT NULL DEFAULT 'terkirim',
    disposed_by_employee_id  UUID               REFERENCES employees(id),
    disposed_at              TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    completed_at             TIMESTAMPTZ
);

CREATE INDEX idx_incoming_status  ON incoming_letters (status, tanggal_terima DESC);
CREATE INDEX idx_incoming_agenda  ON incoming_letters (tanggal_terima DESC);
CREATE INDEX idx_outgoing_year    ON outgoing_letters (tanggal_surat DESC);
CREATE INDEX idx_disposition_assignee ON letter_dispositions (assignee_employee_id, status);
CREATE INDEX idx_disposition_letter   ON letter_dispositions (incoming_letter_id);
