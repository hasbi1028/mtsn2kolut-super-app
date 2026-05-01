-- Sprint 20: dedicated 8 SNP evidence register.
-- Evidence items link documents, programs, performance targets, and owner units into one quality-readiness map.

CREATE TABLE governance_evidence_items (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    period_year             INTEGER     NOT NULL,
    title                   TEXT        NOT NULL,
    evidence_type           TEXT        NOT NULL DEFAULT 'dokumen',
    snp_standard            TEXT        NOT NULL DEFAULT '',
    owner_unit_id           UUID        REFERENCES governance_units(id) ON DELETE SET NULL,
    document_id             UUID        REFERENCES governance_documents(id) ON DELETE SET NULL,
    program_id              UUID        REFERENCES governance_programs(id) ON DELETE SET NULL,
    performance_target_id   UUID        REFERENCES governance_performance_targets(id) ON DELETE SET NULL,
    source_module           TEXT        NOT NULL DEFAULT 'governance',
    evidence_url            TEXT        NOT NULL DEFAULT '',
    status                  TEXT        NOT NULL DEFAULT 'needed',
    notes                   TEXT        NOT NULL DEFAULT '',
    verified_by_user_id     UUID        REFERENCES users(id) ON DELETE SET NULL,
    verified_at             TIMESTAMPTZ,
    created_by_user_id      UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_governance_evidence_year CHECK (period_year >= 2000),
    CONSTRAINT chk_governance_evidence_title CHECK (btrim(title) <> ''),
    CONSTRAINT chk_governance_evidence_type CHECK (evidence_type IN ('dokumen', 'foto', 'tautan', 'laporan', 'arsip', 'lainnya')),
    CONSTRAINT chk_governance_evidence_status CHECK (status IN ('needed', 'collected', 'verified', 'gap')),
    CONSTRAINT chk_governance_evidence_snp CHECK (snp_standard IN ('', 'skl', 'isi', 'proses', 'penilaian', 'ptk', 'sarpras', 'pengelolaan', 'pembiayaan')),
    CONSTRAINT chk_governance_evidence_source CHECK (btrim(source_module) <> '')
);

CREATE INDEX idx_governance_evidence_year_status ON governance_evidence_items(period_year DESC, status);
CREATE INDEX idx_governance_evidence_snp ON governance_evidence_items(snp_standard) WHERE snp_standard <> '';
CREATE INDEX idx_governance_evidence_owner_unit ON governance_evidence_items(owner_unit_id) WHERE owner_unit_id IS NOT NULL;
CREATE INDEX idx_governance_evidence_document ON governance_evidence_items(document_id) WHERE document_id IS NOT NULL;
CREATE INDEX idx_governance_evidence_program ON governance_evidence_items(program_id) WHERE program_id IS NOT NULL;
CREATE INDEX idx_governance_evidence_performance ON governance_evidence_items(performance_target_id) WHERE performance_target_id IS NOT NULL;
