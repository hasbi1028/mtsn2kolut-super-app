-- Sprint 18: Tata Kelola Madrasah foundation
-- Internal governance mapping for organisasi, dokumen, indikator, and 8 SNP evidence.

CREATE TABLE governance_units (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT        NOT NULL UNIQUE,
    name        TEXT        NOT NULL,
    unit_type   TEXT        NOT NULL DEFAULT 'madrasah',
    parent_id   UUID        REFERENCES governance_units(id) ON DELETE SET NULL,
    description TEXT        NOT NULL DEFAULT '',
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order  INTEGER     NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_governance_units_code CHECK (btrim(code) <> ''),
    CONSTRAINT chk_governance_units_name CHECK (btrim(name) <> '')
);

CREATE TABLE governance_positions (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id            UUID        NOT NULL REFERENCES governance_units(id) ON DELETE RESTRICT,
    title              TEXT        NOT NULL,
    position_type      TEXT        NOT NULL DEFAULT 'struktural',
    parent_position_id UUID        REFERENCES governance_positions(id) ON DELETE SET NULL,
    description        TEXT        NOT NULL DEFAULT '',
    tupoksi            TEXT        NOT NULL DEFAULT '',
    is_active          BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order         INTEGER     NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_governance_positions_title CHECK (btrim(title) <> '')
);

CREATE TABLE governance_assignments (
    id                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    position_id              UUID        NOT NULL REFERENCES governance_positions(id) ON DELETE CASCADE,
    employee_id              UUID        NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    start_date               DATE        NOT NULL,
    end_date                 DATE,
    decree_outgoing_letter_id UUID       REFERENCES outgoing_letters(id) ON DELETE SET NULL,
    notes                    TEXT        NOT NULL DEFAULT '',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_governance_assignment_dates CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE UNIQUE INDEX uq_governance_assignment_active_position
    ON governance_assignments(position_id)
    WHERE end_date IS NULL;

CREATE TABLE governance_documents (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    doc_type           TEXT        NOT NULL,
    title              TEXT        NOT NULL,
    period_year        INTEGER     NOT NULL,
    period_label       TEXT        NOT NULL DEFAULT '',
    owner_unit_id      UUID        REFERENCES governance_units(id) ON DELETE SET NULL,
    snp_standard       TEXT        NOT NULL DEFAULT '',
    status             TEXT        NOT NULL DEFAULT 'draft',
    document_url       TEXT        NOT NULL DEFAULT '',
    outgoing_letter_id UUID        REFERENCES outgoing_letters(id) ON DELETE SET NULL,
    summary            TEXT        NOT NULL DEFAULT '',
    created_by_user_id UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_governance_documents_type CHECK (doc_type IN ('visi_misi', 'rkjm', 'rkt', 'renstra', 'perkin', 'iku', 'sk', 'sop', 'snp', 'lainnya')),
    CONSTRAINT chk_governance_documents_status CHECK (status IN ('draft', 'final', 'arsip')),
    CONSTRAINT chk_governance_documents_title CHECK (btrim(title) <> ''),
    CONSTRAINT chk_governance_documents_year CHECK (period_year >= 2000)
);

CREATE TABLE governance_programs (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    period_year             INTEGER     NOT NULL,
    code                    TEXT        NOT NULL,
    name                    TEXT        NOT NULL,
    source_document_id      UUID        REFERENCES governance_documents(id) ON DELETE SET NULL,
    owner_unit_id           UUID        REFERENCES governance_units(id) ON DELETE SET NULL,
    responsible_position_id UUID        REFERENCES governance_positions(id) ON DELETE SET NULL,
    responsible_employee_id UUID        REFERENCES employees(id) ON DELETE SET NULL,
    snp_standard            TEXT        NOT NULL DEFAULT '',
    iku_code                TEXT        NOT NULL DEFAULT '',
    indicator               TEXT        NOT NULL DEFAULT '',
    target_value            TEXT        NOT NULL DEFAULT '',
    target_unit             TEXT        NOT NULL DEFAULT '',
    status                  TEXT        NOT NULL DEFAULT 'planned',
    progress_percent        INTEGER     NOT NULL DEFAULT 0,
    realization_summary     TEXT        NOT NULL DEFAULT '',
    evidence_url            TEXT        NOT NULL DEFAULT '',
    due_date                DATE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_governance_program_year_code UNIQUE (period_year, code),
    CONSTRAINT chk_governance_program_code CHECK (btrim(code) <> ''),
    CONSTRAINT chk_governance_program_name CHECK (btrim(name) <> ''),
    CONSTRAINT chk_governance_program_status CHECK (status IN ('planned', 'in_progress', 'done', 'blocked')),
    CONSTRAINT chk_governance_program_progress CHECK (progress_percent >= 0 AND progress_percent <= 100),
    CONSTRAINT chk_governance_program_year CHECK (period_year >= 2000)
);

CREATE INDEX idx_governance_units_parent ON governance_units(parent_id);
CREATE INDEX idx_governance_positions_unit ON governance_positions(unit_id);
CREATE INDEX idx_governance_positions_parent ON governance_positions(parent_position_id);
CREATE INDEX idx_governance_assignments_employee ON governance_assignments(employee_id, start_date DESC);
CREATE INDEX idx_governance_documents_year_type ON governance_documents(period_year DESC, doc_type);
CREATE INDEX idx_governance_documents_snp ON governance_documents(snp_standard) WHERE snp_standard <> '';
CREATE INDEX idx_governance_programs_year_status ON governance_programs(period_year DESC, status);
CREATE INDEX idx_governance_programs_snp ON governance_programs(snp_standard) WHERE snp_standard <> '';
