-- Sprint 43: Document Cycle module.
-- Turns the MTsN document-cycle guide into cataloged, scheduled, monitorable obligations.

CREATE TABLE document_cycle_catalogs (
    id                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code                     TEXT        NOT NULL UNIQUE,
    title                    TEXT        NOT NULL,
    frequency                TEXT        NOT NULL,
    snp_standard             TEXT        NOT NULL DEFAULT '',
    regulation_ref           TEXT        NOT NULL DEFAULT '',
    default_owner_unit_id    UUID        REFERENCES governance_units(id) ON DELETE SET NULL,
    default_responsible_employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    default_verifier_employee_id    UUID REFERENCES employees(id) ON DELETE SET NULL,
    deadline_days_after_period      INTEGER NOT NULL DEFAULT 5,
    reminder_days_before_due        INTEGER NOT NULL DEFAULT 3,
    description              TEXT        NOT NULL DEFAULT '',
    is_active                BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order               INTEGER     NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_document_cycle_catalog_code CHECK (btrim(code) <> ''),
    CONSTRAINT chk_document_cycle_catalog_title CHECK (btrim(title) <> ''),
    CONSTRAINT chk_document_cycle_catalog_frequency CHECK (frequency IN ('daily', 'weekly', 'monthly', 'quarterly', 'semester', 'annual', 'four_year', 'five_year')),
    CONSTRAINT chk_document_cycle_catalog_deadline CHECK (deadline_days_after_period >= 0 AND deadline_days_after_period <= 365),
    CONSTRAINT chk_document_cycle_catalog_reminder CHECK (reminder_days_before_due >= 0 AND reminder_days_before_due <= 60)
);

CREATE TABLE document_cycle_obligations (
    id                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    catalog_id               UUID        NOT NULL REFERENCES document_cycle_catalogs(id) ON DELETE RESTRICT,
    period_year              INTEGER     NOT NULL,
    period_label             TEXT        NOT NULL,
    period_start             DATE        NOT NULL,
    period_end               DATE        NOT NULL,
    due_date                 DATE        NOT NULL,
    reminder_date            DATE        NOT NULL,
    owner_unit_id            UUID        REFERENCES governance_units(id) ON DELETE SET NULL,
    responsible_employee_id  UUID        REFERENCES employees(id) ON DELETE SET NULL,
    verifier_employee_id     UUID        REFERENCES employees(id) ON DELETE SET NULL,
    status                   TEXT        NOT NULL DEFAULT 'not_started',
    governance_document_id   UUID        REFERENCES governance_documents(id) ON DELETE SET NULL,
    evidence_item_id         UUID        REFERENCES governance_evidence_items(id) ON DELETE SET NULL,
    archive_document_id      UUID        REFERENCES archive_documents(id) ON DELETE SET NULL,
    notes                    TEXT        NOT NULL DEFAULT '',
    verification_notes       TEXT        NOT NULL DEFAULT '',
    completed_at             TIMESTAMPTZ,
    created_by_user_id       UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_document_cycle_obligation_period UNIQUE (catalog_id, period_year, period_label),
    CONSTRAINT chk_document_cycle_obligation_year CHECK (period_year >= 2000),
    CONSTRAINT chk_document_cycle_obligation_period CHECK (period_end >= period_start),
    CONSTRAINT chk_document_cycle_obligation_status CHECK (status IN ('not_started', 'draft', 'waiting_verification', 'completed')),
    CONSTRAINT chk_document_cycle_obligation_completed CHECK ((status = 'completed') OR completed_at IS NULL)
);

CREATE TABLE document_cycle_events (
    id                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    obligation_id            UUID        NOT NULL REFERENCES document_cycle_obligations(id) ON DELETE CASCADE,
    event_type               TEXT        NOT NULL,
    from_status              TEXT        NOT NULL DEFAULT '',
    to_status                TEXT        NOT NULL DEFAULT '',
    notes                    TEXT        NOT NULL DEFAULT '',
    actor_user_id            UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_document_cycle_event_type CHECK (event_type IN ('created', 'updated', 'status_changed', 'generated', 'monitoring_note'))
);

CREATE INDEX idx_document_cycle_catalogs_frequency ON document_cycle_catalogs(frequency, is_active, sort_order);
CREATE INDEX idx_document_cycle_catalogs_snp ON document_cycle_catalogs(snp_standard) WHERE snp_standard <> '';
CREATE INDEX idx_document_cycle_obligations_year_status ON document_cycle_obligations(period_year DESC, status, due_date);
CREATE INDEX idx_document_cycle_obligations_reminder ON document_cycle_obligations(reminder_date, due_date) WHERE status <> 'completed';
CREATE INDEX idx_document_cycle_obligations_owner ON document_cycle_obligations(owner_unit_id, period_year DESC) WHERE owner_unit_id IS NOT NULL;
CREATE INDEX idx_document_cycle_obligations_responsible ON document_cycle_obligations(responsible_employee_id, period_year DESC) WHERE responsible_employee_id IS NOT NULL;
CREATE INDEX idx_document_cycle_events_obligation ON document_cycle_events(obligation_id, created_at DESC);

INSERT INTO document_cycle_catalogs
    (code, title, frequency, snp_standard, regulation_ref, deadline_days_after_period, reminder_days_before_due, description, sort_order)
VALUES
    ('H-DHS', 'Daftar Hadir Siswa per Kelas', 'daily', 'proses', 'Standar Proses', 2, 1, 'Bundel harian absensi siswa sebagai dasar rekap mingguan/bulanan.', 10),
    ('H-DHG', 'Absensi Guru dan Pegawai', 'daily', 'ptk', 'SPTK + SKP BKN', 2, 1, 'Bundel kehadiran pegawai sebagai bukti dukung SKP/e-Kinerja.', 20),
    ('H-JML', 'Jurnal Pembelajaran / Log Mengajar', 'daily', 'proses', 'Standar Proses + SKP', 2, 1, 'Catatan pelaksanaan KBM harian guru mapel.', 30),
    ('H-BPH', 'Buku Piket Harian Guru', 'daily', 'proses', 'Standar Proses', 2, 1, 'Rekam kondisi KBM dan kejadian khusus harian.', 40),
    ('H-BKS', 'Buku Kas Harian / Pengeluaran Harian', 'daily', 'pembiayaan', 'Standar Pembiayaan', 2, 1, 'Catatan transaksi harian sebagai dasar laporan keuangan.', 50),
    ('M-ABS-SISWA', 'Rekap Absensi Siswa Mingguan', 'weekly', 'proses', 'Standar Proses', 2, 1, 'Rekap wali kelas untuk Wakasis setiap akhir pekan.', 110),
    ('M-ABS-PEGAWAI', 'Rekap Kehadiran Guru dan Pegawai Mingguan', 'weekly', 'ptk', 'SPTK + e-Kinerja', 2, 1, 'Monitoring kedisiplinan pegawai pekanan.', 120),
    ('M-PIKET', 'Laporan Piket Mingguan', 'weekly', 'proses', 'Standar Proses', 2, 1, 'Rangkuman kejadian dan kondisi pembelajaran selama sepekan.', 130),
    ('M-RAPAT', 'Agenda / Notulen Rapat Mingguan', 'weekly', 'pengelolaan', 'Standar Pengelolaan', 2, 1, 'Catatan koordinasi kepala, TU, dan unsur madrasah.', 140),
    ('B-ABS-SISWA', 'Rekap Absensi Siswa Bulanan', 'monthly', 'proses', 'Standar Proses', 1, 3, 'Rekap bulanan wali kelas ke Wakasis.', 210),
    ('B-ABS-PEGAWAI', 'Rekap Absensi Guru dan Pegawai Bulanan', 'monthly', 'ptk', 'SPTK + e-Kinerja', 5, 3, 'Rekap bulanan TU untuk kepala/Kankemenag.', 220),
    ('B-KEGIATAN', 'Laporan Kegiatan Bulanan Madrasah', 'monthly', 'pengelolaan', 'Standar Pengelolaan', 10, 3, 'Laporan kegiatan madrasah kepada atasan/instansi.', 230),
    ('B-BOS', 'Laporan Realisasi BOS Bulanan', 'monthly', 'pembiayaan', 'Standar Pembiayaan + RKAM', 10, 3, 'Laporan realisasi anggaran bulanan.', 240),
    ('B-SKP', 'Rekap Input SKP / e-Kinerja Bulanan', 'monthly', 'ptk', 'SKP BKN', 5, 3, 'Monitoring realisasi SKP ASN bulanan.', 250),
    ('B-EMIS', 'Sinkronisasi Data EMIS Bulanan', 'monthly', 'pengelolaan', 'EMIS Kemenag', 10, 3, 'Kontrol update data EMIS sesuai jadwal.', 260),
    ('TW-PERKIN', 'Laporan Triwulan Perjanjian Kinerja', 'quarterly', 'pengelolaan', 'Perkin Kemenag', 10, 7, 'Capaian Perkin kepala madrasah per triwulan.', 310),
    ('TW-SKP', 'Monitoring Realisasi SKP Triwulanan', 'quarterly', 'ptk', 'SKP BKN', 10, 7, 'Monitoring realisasi kuantitas, kualitas, waktu, dan bukti dukung.', 320),
    ('TW-RKAM', 'Laporan Realisasi RKAM Triwulanan', 'quarterly', 'pembiayaan', 'RKAM + BOS', 10, 7, 'Realisasi anggaran per triwulan.', 330),
    ('TW-IKU', 'Evaluasi Capaian IKU Triwulanan', 'quarterly', 'pengelolaan', 'IKU Kemenag', 10, 7, 'Evaluasi indikator kinerja utama madrasah.', 340),
    ('S-RAPOR', 'Rapor Semester', 'semester', 'penilaian', 'Standar Penilaian + SKL', 14, 7, 'Dokumen rapor semester dan rekap nilai akhir.', 410),
    ('S-BK', 'Laporan Program BK Semester', 'semester', 'skl', 'SKL', 14, 7, 'Laporan program dan realisasi layanan BK semester.', 420),
    ('S-SUPERVISI', 'Laporan Supervisi Akademik Semester', 'semester', 'proses', 'Standar Proses + SKP', 14, 7, 'Bukti pelaksanaan supervisi kepala madrasah.', 430),
    ('T-SKP-AWAL', 'Penetapan SKP Semua ASN', 'annual', 'ptk', 'PerBKN No.6/2022, PP No.30/2019', 31, 14, 'Penetapan SKP awal tahun.', 510),
    ('T-PERKIN', 'Penandatanganan Perjanjian Kinerja', 'annual', 'pengelolaan', 'PermenPAN-RB No.53/2014', 31, 14, 'Perkin kepala madrasah cascading dari Kankemenag.', 520),
    ('T-RKT', 'Rencana Kerja Tahunan Madrasah', 'annual', 'pengelolaan', 'Renstra + RKJM + Perkin', 31, 14, 'RKT madrasah sebagai dasar pelaksanaan program tahunan.', 530),
    ('T-RKAM', 'Rencana Kerja dan Anggaran Madrasah', 'annual', 'pembiayaan', 'RKAM + Juknis BOS', 31, 14, 'Perencanaan anggaran tahunan madrasah.', 540),
    ('T-LAKIP', 'LAKIP / LKj Madrasah', 'annual', 'pengelolaan', 'PermenPAN-RB No.53/2014', 31, 14, 'Laporan kinerja akhir tahun kepada Kankemenag.', 550),
    ('T-EDM', 'EDS / EDM Tahunan', 'annual', 'pengelolaan', '8 SNP + Renstra', 31, 14, 'Evaluasi diri madrasah untuk dasar RKT berikutnya.', 560),
    ('RKJM', 'Dokumen RKJM 4 Tahunan', 'four_year', 'pengelolaan', '8 SNP + Renstra', 120, 30, 'Rencana kerja jangka menengah periode 4 tahun.', 610),
    ('RENSTRA', 'Dokumen Renstra MTsN 5 Tahunan', 'five_year', 'pengelolaan', 'Renstra Kemenag + RPJMN + 8 SNP', 180, 45, 'Dokumen perencanaan strategis madrasah 5 tahunan.', 710)
ON CONFLICT (code) DO UPDATE SET
    title = EXCLUDED.title,
    frequency = EXCLUDED.frequency,
    snp_standard = EXCLUDED.snp_standard,
    regulation_ref = EXCLUDED.regulation_ref,
    deadline_days_after_period = EXCLUDED.deadline_days_after_period,
    reminder_days_before_due = EXCLUDED.reminder_days_before_due,
    description = EXCLUDED.description,
    sort_order = EXCLUDED.sort_order,
    is_active = TRUE,
    updated_at = NOW();
