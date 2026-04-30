-- Migration 034: Jurnal Kelas (Class Journal)
-- Records daily teaching sessions with student attendance per subject per class.

CREATE TYPE journal_attendance_status AS ENUM ('hadir', 'sakit', 'izin', 'alpha');

-- One row per teaching session (one assignment on one date)
CREATE TABLE class_journal_sessions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id   UUID        NOT NULL REFERENCES class_subject_assignments(id) ON DELETE CASCADE,
    tanggal         DATE        NOT NULL,
    pertemuan_ke    INT         NOT NULL,
    materi          TEXT        NOT NULL DEFAULT '',
    kegiatan        TEXT        NOT NULL DEFAULT '',
    catatan         TEXT        NOT NULL DEFAULT '',
    guru_hadir      BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_journal_session_date     UNIQUE (assignment_id, tanggal),
    CONSTRAINT uq_journal_session_meeting  UNIQUE (assignment_id, pertemuan_ke),
    CONSTRAINT chk_pertemuan_ke            CHECK  (pertemuan_ke >= 1)
);

CREATE INDEX idx_journal_sessions_assignment
    ON class_journal_sessions (assignment_id, tanggal DESC);

-- One row per student per session
CREATE TABLE class_journal_attendances (
    id          UUID                      PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID                      NOT NULL REFERENCES class_journal_sessions(id) ON DELETE CASCADE,
    student_id  UUID                      NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    status      journal_attendance_status NOT NULL DEFAULT 'hadir',
    catatan     TEXT                      NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ               NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ               NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_journal_attendance UNIQUE (session_id, student_id)
);

CREATE INDEX idx_journal_attendances_session
    ON class_journal_attendances (session_id, student_id);

CREATE INDEX idx_journal_attendances_student
    ON class_journal_attendances (student_id);
