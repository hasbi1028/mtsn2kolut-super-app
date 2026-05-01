-- Sprint 23: Kesiswaan engagement, BK, and transfer administration.

CREATE TABLE extracurriculars (
    id                     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code                   TEXT        NOT NULL UNIQUE,
    name                   TEXT        NOT NULL,
    category               TEXT        NOT NULL DEFAULT '',
    description            TEXT        NOT NULL DEFAULT '',
    supervisor_employee_id UUID        REFERENCES employees(id) ON DELETE SET NULL,
    schedule_text          TEXT        NOT NULL DEFAULT '',
    is_active              BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_extracurriculars_code CHECK (btrim(code) <> ''),
    CONSTRAINT chk_extracurriculars_name CHECK (btrim(name) <> '')
);

CREATE TABLE extracurricular_members (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    extracurricular_id  UUID        NOT NULL REFERENCES extracurriculars(id) ON DELETE CASCADE,
    student_id          UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    joined_at           DATE        NOT NULL DEFAULT CURRENT_DATE,
    role                TEXT        NOT NULL DEFAULT 'member',
    status              TEXT        NOT NULL DEFAULT 'active',
    notes               TEXT        NOT NULL DEFAULT '',
    recorded_by_user_id UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_extracurricular_members_student UNIQUE (extracurricular_id, student_id),
    CONSTRAINT chk_extracurricular_members_role CHECK (role IN ('member', 'leader', 'assistant')),
    CONSTRAINT chk_extracurricular_members_status CHECK (status IN ('active', 'inactive', 'alumni'))
);

CREATE TABLE counseling_sessions (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id            UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    session_date          DATE        NOT NULL DEFAULT CURRENT_DATE,
    topic                 TEXT        NOT NULL,
    summary               TEXT        NOT NULL DEFAULT '',
    follow_up             TEXT        NOT NULL DEFAULT '',
    status                TEXT        NOT NULL DEFAULT 'open',
    is_confidential       BOOLEAN     NOT NULL DEFAULT FALSE,
    counselor_employee_id UUID        REFERENCES employees(id) ON DELETE SET NULL,
    recorded_by_user_id   UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_counseling_sessions_topic CHECK (btrim(topic) <> ''),
    CONSTRAINT chk_counseling_sessions_status CHECK (status IN ('open', 'monitoring', 'resolved', 'referred', 'canceled'))
);

CREATE TABLE student_transfers (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id          UUID        NOT NULL REFERENCES students(id) ON DELETE RESTRICT,
    transfer_date       DATE        NOT NULL DEFAULT CURRENT_DATE,
    transfer_type       TEXT        NOT NULL,
    previous_school     TEXT        NOT NULL DEFAULT '',
    destination_school  TEXT        NOT NULL DEFAULT '',
    reason              TEXT        NOT NULL DEFAULT '',
    document_ref        TEXT        NOT NULL DEFAULT '',
    notes               TEXT        NOT NULL DEFAULT '',
    status              TEXT        NOT NULL DEFAULT 'completed',
    recorded_by_user_id UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_student_transfers_type CHECK (transfer_type IN ('in', 'out')),
    CONSTRAINT chk_student_transfers_status CHECK (status IN ('completed', 'canceled')),
    CONSTRAINT chk_student_transfers_destination CHECK (transfer_type <> 'out' OR btrim(destination_school) <> ''),
    CONSTRAINT chk_student_transfers_previous CHECK (transfer_type <> 'in' OR btrim(previous_school) <> '')
);

CREATE INDEX idx_extracurriculars_active ON extracurriculars(is_active, name);
CREATE INDEX idx_extracurricular_members_activity ON extracurricular_members(extracurricular_id, status);
CREATE INDEX idx_extracurricular_members_student ON extracurricular_members(student_id, status);
CREATE INDEX idx_counseling_sessions_student_date ON counseling_sessions(student_id, session_date DESC);
CREATE INDEX idx_counseling_sessions_status ON counseling_sessions(status, session_date DESC);
CREATE INDEX idx_counseling_sessions_confidential ON counseling_sessions(is_confidential, status);
CREATE INDEX idx_student_transfers_student_date ON student_transfers(student_id, transfer_date DESC);
CREATE INDEX idx_student_transfers_type ON student_transfers(transfer_type, transfer_date DESC);
