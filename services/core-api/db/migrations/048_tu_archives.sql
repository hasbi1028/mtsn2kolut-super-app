-- Sprint 24: Tata Usaha archive register and file storage metadata.

CREATE TABLE archive_categories (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code                TEXT        NOT NULL UNIQUE,
    name                TEXT        NOT NULL,
    classification_code TEXT        REFERENCES letter_classifications(code) ON DELETE SET NULL,
    description         TEXT        NOT NULL DEFAULT '',
    retention_years     INTEGER     NOT NULL DEFAULT 5,
    is_active           BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_archive_categories_code CHECK (btrim(code) <> ''),
    CONSTRAINT chk_archive_categories_name CHECK (btrim(name) <> ''),
    CONSTRAINT chk_archive_categories_retention CHECK (retention_years >= 0)
);

CREATE TABLE archive_documents (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id         UUID        NOT NULL REFERENCES archive_categories(id) ON DELETE RESTRICT,
    title               TEXT        NOT NULL,
    archive_number      TEXT        NOT NULL DEFAULT '',
    document_date       DATE,
    received_date       DATE        NOT NULL DEFAULT CURRENT_DATE,
    summary             TEXT        NOT NULL DEFAULT '',
    tags                TEXT        NOT NULL DEFAULT '',
    status              TEXT        NOT NULL DEFAULT 'active',
    storage_location    TEXT        NOT NULL DEFAULT '',
    retention_until     DATE,
    original_name       TEXT        NOT NULL,
    stored_name         TEXT        NOT NULL UNIQUE,
    file_path           TEXT        NOT NULL,
    mime_type           TEXT        NOT NULL,
    file_size           BIGINT      NOT NULL,
    checksum_sha256     TEXT        NOT NULL,
    uploaded_by_user_id UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_archive_documents_title CHECK (btrim(title) <> ''),
    CONSTRAINT chk_archive_documents_status CHECK (status IN ('active', 'borrowed', 'disposed')),
    CONSTRAINT chk_archive_documents_file_size CHECK (file_size > 0),
    CONSTRAINT chk_archive_documents_file_path CHECK (btrim(file_path) <> '')
);

CREATE UNIQUE INDEX uq_archive_documents_number_not_empty
    ON archive_documents(archive_number)
    WHERE archive_number <> '';

CREATE INDEX idx_archive_categories_active ON archive_categories(is_active, code);
CREATE INDEX idx_archive_documents_category ON archive_documents(category_id, received_date DESC);
CREATE INDEX idx_archive_documents_status ON archive_documents(status, received_date DESC);
CREATE INDEX idx_archive_documents_document_date ON archive_documents(document_date DESC) WHERE document_date IS NOT NULL;
CREATE INDEX idx_archive_documents_retention ON archive_documents(retention_until) WHERE retention_until IS NOT NULL;

INSERT INTO archive_categories (code, name, classification_code, description, retention_years) VALUES
    ('ARS-SM', 'Surat Masuk', 'OT.00', 'Arsip surat masuk dan bukti penerimaan surat.', 5),
    ('ARS-SK', 'Surat Keluar', 'OT.00', 'Arsip surat keluar, surat keterangan, undangan, dan surat balasan.', 5),
    ('ARS-KS', 'Kesiswaan', 'PP.00.2', 'Arsip administrasi peserta didik dan kegiatan kesiswaan.', 5),
    ('ARS-KP', 'Kepegawaian', 'KP.00', 'Arsip administrasi kepegawaian madrasah.', 10),
    ('ARS-KU', 'Keuangan', 'KU.00', 'Arsip administrasi keuangan, anggaran, dan pertanggungjawaban.', 10),
    ('ARS-SAPRAS', 'Sarana Prasarana', 'KS.00', 'Arsip inventaris, pemeliharaan, dan sarana prasarana.', 10)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    classification_code = EXCLUDED.classification_code,
    description = EXCLUDED.description,
    retention_years = EXCLUDED.retention_years,
    is_active = TRUE,
    updated_at = NOW();
