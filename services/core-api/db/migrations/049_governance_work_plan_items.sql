-- Sprint 25: RKT/RKJM execution items.
-- Breaks governance programs into annual work-plan activities with budget, schedule, evidence, and progress tracking.

CREATE TABLE governance_work_plan_items (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    period_year             INTEGER     NOT NULL,
    program_id              UUID        NOT NULL REFERENCES governance_programs(id) ON DELETE CASCADE,
    source_document_id      UUID        REFERENCES governance_documents(id) ON DELETE SET NULL,
    owner_unit_id           UUID        REFERENCES governance_units(id) ON DELETE SET NULL,
    responsible_employee_id UUID        REFERENCES employees(id) ON DELETE SET NULL,
    evidence_item_id        UUID        REFERENCES governance_evidence_items(id) ON DELETE SET NULL,
    activity_code           TEXT        NOT NULL,
    activity_name           TEXT        NOT NULL,
    output_indicator        TEXT        NOT NULL DEFAULT '',
    target_volume           TEXT        NOT NULL DEFAULT '',
    target_unit             TEXT        NOT NULL DEFAULT '',
    budget_source           TEXT        NOT NULL DEFAULT '',
    budget_amount           BIGINT      NOT NULL DEFAULT 0,
    realization_amount      BIGINT      NOT NULL DEFAULT 0,
    status                  TEXT        NOT NULL DEFAULT 'planned',
    progress_percent        INTEGER     NOT NULL DEFAULT 0,
    start_date              DATE,
    end_date                DATE,
    evidence_url            TEXT        NOT NULL DEFAULT '',
    notes                   TEXT        NOT NULL DEFAULT '',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_governance_work_plan_program_code UNIQUE (program_id, activity_code),
    CONSTRAINT chk_governance_work_plan_year CHECK (period_year >= 2000),
    CONSTRAINT chk_governance_work_plan_code CHECK (btrim(activity_code) <> ''),
    CONSTRAINT chk_governance_work_plan_name CHECK (btrim(activity_name) <> ''),
    CONSTRAINT chk_governance_work_plan_budget CHECK (budget_amount >= 0),
    CONSTRAINT chk_governance_work_plan_realization CHECK (realization_amount >= 0),
    CONSTRAINT chk_governance_work_plan_status CHECK (status IN ('planned', 'in_progress', 'done', 'blocked')),
    CONSTRAINT chk_governance_work_plan_progress CHECK (progress_percent >= 0 AND progress_percent <= 100),
    CONSTRAINT chk_governance_work_plan_dates CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_governance_work_plan_year_status ON governance_work_plan_items(period_year DESC, status);
CREATE INDEX idx_governance_work_plan_program ON governance_work_plan_items(program_id, activity_code);
CREATE INDEX idx_governance_work_plan_owner_unit ON governance_work_plan_items(owner_unit_id) WHERE owner_unit_id IS NOT NULL;
CREATE INDEX idx_governance_work_plan_responsible ON governance_work_plan_items(responsible_employee_id) WHERE responsible_employee_id IS NOT NULL;
CREATE INDEX idx_governance_work_plan_evidence ON governance_work_plan_items(evidence_item_id) WHERE evidence_item_id IS NOT NULL;
