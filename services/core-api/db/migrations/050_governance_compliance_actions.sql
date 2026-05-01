-- Sprint 31: Governance compliance action tracker.
-- Turns Renstra/IKU/RKT/SKP/8 SNP mapping gaps into assigned follow-up work.

CREATE TABLE governance_compliance_actions (
    id                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    period_year              INTEGER     NOT NULL,
    source_type              TEXT        NOT NULL DEFAULT 'manual',
    source_ref_id            UUID,
    snp_standard             TEXT        NOT NULL DEFAULT '',
    program_id               UUID        REFERENCES governance_programs(id) ON DELETE SET NULL,
    document_id              UUID        REFERENCES governance_documents(id) ON DELETE SET NULL,
    performance_target_id    UUID        REFERENCES governance_performance_targets(id) ON DELETE SET NULL,
    evidence_item_id         UUID        REFERENCES governance_evidence_items(id) ON DELETE SET NULL,
    owner_unit_id            UUID        REFERENCES governance_units(id) ON DELETE SET NULL,
    responsible_employee_id  UUID        REFERENCES employees(id) ON DELETE SET NULL,
    title                    TEXT        NOT NULL,
    description              TEXT        NOT NULL DEFAULT '',
    priority                 TEXT        NOT NULL DEFAULT 'medium',
    status                   TEXT        NOT NULL DEFAULT 'open',
    due_date                 DATE,
    completed_at             TIMESTAMPTZ,
    follow_up_notes          TEXT        NOT NULL DEFAULT '',
    evidence_url             TEXT        NOT NULL DEFAULT '',
    created_by_user_id       UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_governance_compliance_action_year CHECK (period_year >= 2000),
    CONSTRAINT chk_governance_compliance_action_source CHECK (source_type IN ('alignment_gap', 'snp_gap', 'audit', 'document', 'program', 'performance', 'evidence', 'manual')),
    CONSTRAINT chk_governance_compliance_action_title CHECK (btrim(title) <> ''),
    CONSTRAINT chk_governance_compliance_action_priority CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    CONSTRAINT chk_governance_compliance_action_status CHECK (status IN ('open', 'in_progress', 'waiting_evidence', 'done', 'cancelled')),
    CONSTRAINT chk_governance_compliance_action_done_time CHECK ((status = 'done') OR completed_at IS NULL)
);

CREATE INDEX idx_governance_compliance_actions_year_status ON governance_compliance_actions(period_year DESC, status);
CREATE INDEX idx_governance_compliance_actions_priority_due ON governance_compliance_actions(priority, due_date) WHERE status NOT IN ('done', 'cancelled');
CREATE INDEX idx_governance_compliance_actions_snp ON governance_compliance_actions(snp_standard) WHERE snp_standard <> '';
CREATE INDEX idx_governance_compliance_actions_program ON governance_compliance_actions(program_id) WHERE program_id IS NOT NULL;
CREATE INDEX idx_governance_compliance_actions_responsible ON governance_compliance_actions(responsible_employee_id) WHERE responsible_employee_id IS NOT NULL;
