-- Sprint 44: Document-cycle integration spine.
-- Adds bidang/external-system classification and deeper links into governance execution records.

ALTER TABLE document_cycle_catalogs
    ADD COLUMN domain_area TEXT NOT NULL DEFAULT 'governance',
    ADD COLUMN external_system TEXT NOT NULL DEFAULT '';

ALTER TABLE document_cycle_obligations
    ADD COLUMN domain_area TEXT NOT NULL DEFAULT 'governance',
    ADD COLUMN external_system TEXT NOT NULL DEFAULT '',
    ADD COLUMN work_plan_item_id UUID REFERENCES governance_work_plan_items(id) ON DELETE SET NULL,
    ADD COLUMN performance_target_id UUID REFERENCES governance_performance_targets(id) ON DELETE SET NULL,
    ADD COLUMN compliance_action_id UUID REFERENCES governance_compliance_actions(id) ON DELETE SET NULL;

ALTER TABLE document_cycle_catalogs
    ADD CONSTRAINT chk_document_cycle_catalog_domain_area
        CHECK (domain_area IN ('tu', 'kesiswaan', 'kurikulum', 'sarpras', 'governance', 'keuangan', 'eksternal')),
    ADD CONSTRAINT chk_document_cycle_catalog_external_system
        CHECK (external_system IN ('', 'skp_bkn', 'emis', 'sipka', 'simak_bmn', 'rkam_bos', 'perkin', 'iku', 'lakip_lkj', 'edm'));

ALTER TABLE document_cycle_obligations
    ADD CONSTRAINT chk_document_cycle_obligation_domain_area
        CHECK (domain_area IN ('tu', 'kesiswaan', 'kurikulum', 'sarpras', 'governance', 'keuangan', 'eksternal')),
    ADD CONSTRAINT chk_document_cycle_obligation_external_system
        CHECK (external_system IN ('', 'skp_bkn', 'emis', 'sipka', 'simak_bmn', 'rkam_bos', 'perkin', 'iku', 'lakip_lkj', 'edm'));

UPDATE document_cycle_catalogs
SET
    domain_area = CASE
        WHEN code IN ('H-DHS', 'M-ABS-SISWA', 'B-ABS-SISWA', 'S-BK') THEN 'kesiswaan'
        WHEN code IN ('H-JML', 'H-BPH', 'S-RAPOR', 'S-SUPERVISI') THEN 'kurikulum'
        WHEN code IN ('H-BKS', 'B-BOS', 'TW-RKAM', 'T-RKAM') THEN 'keuangan'
        WHEN code IN ('B-SKP', 'TW-SKP', 'T-SKP-AWAL', 'B-EMIS', 'TW-PERKIN', 'TW-IKU', 'T-PERKIN', 'T-LAKIP') THEN 'eksternal'
        WHEN code IN ('H-DHG', 'M-ABS-PEGAWAI', 'B-ABS-PEGAWAI') THEN 'tu'
        ELSE 'governance'
    END,
    external_system = CASE
        WHEN code IN ('B-SKP', 'TW-SKP', 'T-SKP-AWAL') THEN 'skp_bkn'
        WHEN code = 'B-EMIS' THEN 'emis'
        WHEN code IN ('H-BKS', 'B-BOS', 'TW-RKAM', 'T-RKAM') THEN 'rkam_bos'
        WHEN code IN ('TW-PERKIN', 'T-PERKIN') THEN 'perkin'
        WHEN code = 'TW-IKU' THEN 'iku'
        WHEN code = 'T-LAKIP' THEN 'lakip_lkj'
        WHEN code = 'T-EDM' THEN 'edm'
        ELSE ''
    END,
    updated_at = NOW();

UPDATE document_cycle_obligations o
SET
    domain_area = c.domain_area,
    external_system = c.external_system,
    updated_at = NOW()
FROM document_cycle_catalogs c
WHERE c.id = o.catalog_id;

INSERT INTO document_cycle_catalogs
    (code, title, frequency, snp_standard, regulation_ref, domain_area, external_system, deadline_days_after_period, reminder_days_before_due, description, sort_order)
VALUES
    ('B-SIPKA', 'Update SIPKA Bulanan', 'monthly', 'pengelolaan', 'SIPKA Kemenag', 'eksternal', 'sipka', 10, 3, 'Checklist internal untuk memastikan input/validasi SIPKA dilakukan di portal resmi.', 270),
    ('TW-SIMAK-BMN', 'Rekonsiliasi SIMAK-BMN Triwulanan', 'quarterly', 'sarpras', 'SIMAK-BMN / BMN Kemenag', 'sarpras', 'simak_bmn', 10, 7, 'Checklist bukti aset dan rekonsiliasi SIMAK-BMN; aplikasi ini hanya menyimpan status dan bukti dukung.', 350)
ON CONFLICT (code) DO UPDATE SET
    title = EXCLUDED.title,
    frequency = EXCLUDED.frequency,
    snp_standard = EXCLUDED.snp_standard,
    regulation_ref = EXCLUDED.regulation_ref,
    domain_area = EXCLUDED.domain_area,
    external_system = EXCLUDED.external_system,
    deadline_days_after_period = EXCLUDED.deadline_days_after_period,
    reminder_days_before_due = EXCLUDED.reminder_days_before_due,
    description = EXCLUDED.description,
    sort_order = EXCLUDED.sort_order,
    is_active = TRUE,
    updated_at = NOW();

CREATE INDEX idx_document_cycle_catalogs_domain_area ON document_cycle_catalogs(domain_area, is_active, sort_order);
CREATE INDEX idx_document_cycle_catalogs_external_system ON document_cycle_catalogs(external_system, is_active) WHERE external_system <> '';
CREATE INDEX idx_document_cycle_obligations_domain_area ON document_cycle_obligations(domain_area, period_year DESC, status);
CREATE INDEX idx_document_cycle_obligations_external_system ON document_cycle_obligations(external_system, period_year DESC, status) WHERE external_system <> '';
CREATE INDEX idx_document_cycle_obligations_work_plan ON document_cycle_obligations(work_plan_item_id) WHERE work_plan_item_id IS NOT NULL;
CREATE INDEX idx_document_cycle_obligations_performance_target ON document_cycle_obligations(performance_target_id) WHERE performance_target_id IS NOT NULL;
CREATE INDEX idx_document_cycle_obligations_compliance_action ON document_cycle_obligations(compliance_action_id) WHERE compliance_action_id IS NOT NULL;
