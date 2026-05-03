CREATE TABLE school_rooms (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code             TEXT        NOT NULL UNIQUE,
    name             TEXT        NOT NULL,
    building         TEXT        NOT NULL DEFAULT '',
    floor            TEXT        NOT NULL DEFAULT '',
    room_type        TEXT        NOT NULL DEFAULT 'kelas',
    location_note    TEXT        NOT NULL DEFAULT '',
    default_capacity INTEGER     NOT NULL DEFAULT 30,
    exam_capacity    INTEGER     NOT NULL DEFAULT 30,
    condition        TEXT        NOT NULL DEFAULT 'baik',
    is_exam_eligible BOOLEAN     NOT NULL DEFAULT TRUE,
    network_ready    BOOLEAN     NOT NULL DEFAULT FALSE,
    power_ready      BOOLEAN     NOT NULL DEFAULT FALSE,
    notes            TEXT        NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_school_rooms_condition CHECK (condition IN ('baik', 'perlu-perawatan', 'rusak')),
    CONSTRAINT chk_school_rooms_default_capacity CHECK (default_capacity >= 1),
    CONSTRAINT chk_school_rooms_exam_capacity CHECK (exam_capacity >= 1)
);

CREATE INDEX idx_school_rooms_exam_eligible ON school_rooms(is_exam_eligible);
CREATE INDEX idx_school_rooms_type          ON school_rooms(room_type);
CREATE INDEX idx_school_rooms_condition     ON school_rooms(condition);

ALTER TABLE cbt_exam_rooms
    ADD COLUMN school_room_id    UUID REFERENCES school_rooms(id) ON DELETE SET NULL,
    ADD COLUMN room_name_snapshot TEXT NOT NULL DEFAULT '',
    ADD COLUMN capacity_override INTEGER,
    ADD COLUMN room_token        TEXT NOT NULL DEFAULT '',
    ADD COLUMN status            TEXT NOT NULL DEFAULT 'draft',
    ADD COLUMN is_locked         BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD CONSTRAINT chk_cbt_exam_rooms_capacity_override CHECK (capacity_override IS NULL OR capacity_override >= 1),
    ADD CONSTRAINT chk_cbt_exam_rooms_status CHECK (status IN ('draft', 'ready', 'active', 'locked', 'closed'));

UPDATE cbt_exam_rooms
SET room_name_snapshot = room_name
WHERE room_name_snapshot = '';

CREATE INDEX idx_cbt_exam_rooms_school_room ON cbt_exam_rooms(school_room_id);
CREATE UNIQUE INDEX idx_cbt_exam_rooms_token ON cbt_exam_rooms(room_token) WHERE room_token <> '';

CREATE TABLE cbt_room_proctors (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_room_id UUID        NOT NULL REFERENCES cbt_exam_rooms(id) ON DELETE CASCADE,
    employee_id  UUID        NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    role         TEXT        NOT NULL DEFAULT 'pendamping',
    assigned_by  UUID        REFERENCES users(id) ON DELETE SET NULL,
    assigned_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_cbt_room_proctors_role CHECK (role IN ('utama', 'pendamping', 'cadangan')),
    CONSTRAINT uq_cbt_room_proctors_room_employee UNIQUE (exam_room_id, employee_id)
);

CREATE INDEX idx_cbt_room_proctors_room     ON cbt_room_proctors(exam_room_id);
CREATE INDEX idx_cbt_room_proctors_employee ON cbt_room_proctors(employee_id);
