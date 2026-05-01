-- Sprint 19: SKP mirror and internal performance-target cascading.
-- This is an internal planning/evidence surface, not a replacement for BKN/SIPKA.

CREATE TABLE governance_performance_targets (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    period_year             INTEGER     NOT NULL,
    employee_id             UUID        NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    position_id             UUID        REFERENCES governance_positions(id) ON DELETE SET NULL,
    program_id              UUID        REFERENCES governance_programs(id) ON DELETE SET NULL,
    parent_target_id        UUID        REFERENCES governance_performance_targets(id) ON DELETE SET NULL,
    aspect                  TEXT        NOT NULL DEFAULT 'hasil_kerja',
    title                   TEXT        NOT NULL,
    indicator               TEXT        NOT NULL DEFAULT '',
    target_value            TEXT        NOT NULL DEFAULT '',
    target_unit             TEXT        NOT NULL DEFAULT '',
    status                  TEXT        NOT NULL DEFAULT 'planned',
    progress_percent        INTEGER     NOT NULL DEFAULT 0,
    evidence_url            TEXT        NOT NULL DEFAULT '',
    review_notes            TEXT        NOT NULL DEFAULT '',
    due_date                DATE,
    created_by_user_id      UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_governance_perf_year CHECK (period_year >= 2000),
    CONSTRAINT chk_governance_perf_aspect CHECK (aspect IN ('hasil_kerja', 'perilaku_kerja', 'tambahan')),
    CONSTRAINT chk_governance_perf_title CHECK (btrim(title) <> ''),
    CONSTRAINT chk_governance_perf_status CHECK (status IN ('planned', 'in_progress', 'done', 'blocked')),
    CONSTRAINT chk_governance_perf_progress CHECK (progress_percent >= 0 AND progress_percent <= 100)
);

CREATE INDEX idx_governance_perf_employee_year ON governance_performance_targets(employee_id, period_year DESC);
CREATE INDEX idx_governance_perf_program ON governance_performance_targets(program_id) WHERE program_id IS NOT NULL;
CREATE INDEX idx_governance_perf_parent ON governance_performance_targets(parent_target_id) WHERE parent_target_id IS NOT NULL;
CREATE INDEX idx_governance_perf_status ON governance_performance_targets(period_year DESC, status);
