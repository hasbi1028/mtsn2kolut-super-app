CREATE TABLE cbt_room_handovers (
    id                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_room_id             UUID        NOT NULL UNIQUE REFERENCES cbt_exam_rooms(id) ON DELETE CASCADE,
    attendance_checked       BOOLEAN     NOT NULL DEFAULT FALSE,
    all_submitted_checked    BOOLEAN     NOT NULL DEFAULT FALSE,
    device_issue_checked     BOOLEAN     NOT NULL DEFAULT FALSE,
    room_clean_checked       BOOLEAN     NOT NULL DEFAULT FALSE,
    token_returned_checked   BOOLEAN     NOT NULL DEFAULT FALSE,
    assets_returned_checked  BOOLEAN     NOT NULL DEFAULT FALSE,
    incident_notes           TEXT        NOT NULL DEFAULT '',
    operator_notes           TEXT        NOT NULL DEFAULT '',
    handover_notes           TEXT        NOT NULL DEFAULT '',
    locked_at                TIMESTAMPTZ,
    locked_by                UUID        REFERENCES users(id) ON DELETE SET NULL,
    updated_by               UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cbt_room_handovers_locked
    ON cbt_room_handovers(locked_at)
    WHERE locked_at IS NOT NULL;
