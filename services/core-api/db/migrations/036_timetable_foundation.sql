CREATE TABLE timetable_slots (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL REFERENCES class_subject_assignments(id) ON DELETE CASCADE,
    day_of_week   SMALLINT NOT NULL CHECK (day_of_week BETWEEN 1 AND 6),
    start_time    TIME NOT NULL,
    end_time      TIME NOT NULL,
    room_label    TEXT NOT NULL DEFAULT '',
    notes         TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_timetable_slot_time_range CHECK (end_time > start_time)
);

CREATE INDEX idx_timetable_slots_assignment_day
    ON timetable_slots (assignment_id, day_of_week, start_time);

CREATE INDEX idx_timetable_slots_day_time
    ON timetable_slots (day_of_week, start_time);
