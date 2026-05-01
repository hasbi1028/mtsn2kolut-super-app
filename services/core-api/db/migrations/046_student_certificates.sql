-- Sprint 22: Tata Usaha student certificate issuance.
-- Surat keterangan siswa is tied to the central outgoing-letter register.

CREATE TABLE certificate_templates (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code            TEXT        NOT NULL UNIQUE,
    name            TEXT        NOT NULL,
    description     TEXT        NOT NULL DEFAULT '',
    default_purpose TEXT        NOT NULL DEFAULT '',
    body_template   TEXT        NOT NULL DEFAULT '',
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_certificate_templates_code CHECK (btrim(code) <> ''),
    CONSTRAINT chk_certificate_templates_name CHECK (btrim(name) <> '')
);

CREATE TABLE student_certificates (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id         UUID        NOT NULL REFERENCES certificate_templates(id) ON DELETE RESTRICT,
    student_id          UUID        NOT NULL REFERENCES students(id) ON DELETE RESTRICT,
    outgoing_letter_id  UUID        NOT NULL UNIQUE REFERENCES outgoing_letters(id) ON DELETE RESTRICT,
    tanggal_surat       DATE        NOT NULL,
    purpose             TEXT        NOT NULL DEFAULT '',
    recipient           TEXT        NOT NULL DEFAULT 'Yang berkepentingan',
    remarks             TEXT        NOT NULL DEFAULT '',
    snapshot_data       JSONB       NOT NULL DEFAULT '{}'::jsonb,
    status              TEXT        NOT NULL DEFAULT 'issued',
    created_by_user_id  UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_student_certificates_recipient CHECK (btrim(recipient) <> ''),
    CONSTRAINT chk_student_certificates_status CHECK (status IN ('issued', 'canceled'))
);

CREATE INDEX idx_certificate_templates_active ON certificate_templates(is_active, name);
CREATE INDEX idx_student_certificates_student ON student_certificates(student_id, tanggal_surat DESC);
CREATE INDEX idx_student_certificates_template ON student_certificates(template_id, tanggal_surat DESC);
CREATE INDEX idx_student_certificates_status ON student_certificates(status, tanggal_surat DESC);

INSERT INTO certificate_templates (code, name, description, default_purpose, body_template) VALUES
    (
        'aktif',
        'Surat Keterangan Aktif',
        'Keterangan bahwa peserta didik masih aktif belajar di madrasah.',
        'Melengkapi administrasi yang membutuhkan keterangan status aktif peserta didik.',
        'Yang bersangkutan benar tercatat sebagai peserta didik aktif di MTs Negeri 2 Kolaka Utara.'
    ),
    (
        'lulus',
        'Surat Keterangan Lulus',
        'Keterangan kelulusan peserta didik untuk kebutuhan administrasi lanjutan.',
        'Melengkapi administrasi pendaftaran atau verifikasi kelulusan.',
        'Yang bersangkutan benar telah menyelesaikan pendidikan di MTs Negeri 2 Kolaka Utara sesuai data madrasah.'
    ),
    (
        'pindah',
        'Surat Keterangan Pindah',
        'Keterangan perpindahan peserta didik dari atau ke satuan pendidikan lain.',
        'Melengkapi administrasi perpindahan peserta didik.',
        'Yang bersangkutan benar tercatat dalam administrasi peserta didik MTs Negeri 2 Kolaka Utara untuk keperluan perpindahan sekolah.'
    ),
    (
        'kehilangan_dokumen',
        'Surat Keterangan Kehilangan Dokumen',
        'Keterangan pendukung untuk pengurusan dokumen madrasah yang hilang.',
        'Melengkapi administrasi pengurusan ulang dokumen madrasah.',
        'Yang bersangkutan benar tercatat dalam data peserta didik MTs Negeri 2 Kolaka Utara sebagai dasar pengurusan dokumen.'
    ),
    (
        'mengikuti_kegiatan',
        'Surat Keterangan Mengikuti Kegiatan',
        'Keterangan peserta didik untuk mengikuti lomba, seleksi, atau kegiatan resmi.',
        'Melengkapi administrasi keikutsertaan peserta didik dalam kegiatan.',
        'Yang bersangkutan benar merupakan peserta didik MTs Negeri 2 Kolaka Utara dan mendapat keterangan untuk mengikuti kegiatan sebagaimana kebutuhan administrasi.'
    )
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    default_purpose = EXCLUDED.default_purpose,
    body_template = EXCLUDED.body_template,
    is_active = TRUE,
    updated_at = NOW();
