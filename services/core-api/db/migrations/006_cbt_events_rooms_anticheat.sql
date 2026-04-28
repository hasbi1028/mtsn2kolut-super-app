-- Migration 006: CBT Exam Events, Rooms, Anti-cheat

CREATE TYPE cbt_exam_type AS ENUM ('ulangan', 'uts', 'uas', 'uam', 'tryout', 'lainnya');

-- Kegiatan ujian (UTS, UAS, UAM, Ulangan, Try Out)
-- event_id di sessions bersifat nullable: ulangan harian tidak wajib punya event
CREATE TABLE cbt_exam_events (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  title            TEXT        NOT NULL,
  exam_type        cbt_exam_type NOT NULL DEFAULT 'lainnya',
  scope            TEXT        NOT NULL DEFAULT 'class',  -- 'class'|'grade'|'school'
  academic_year_id UUID        REFERENCES academic_years(id) ON DELETE SET NULL,
  status           TEXT        NOT NULL DEFAULT 'draft',  -- draft|active|finished
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ruangan ujian per sesi — kapasitas divalidasi saat shuffle
CREATE TABLE cbt_exam_rooms (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id  UUID        NOT NULL REFERENCES cbt_exam_sessions(id) ON DELETE CASCADE,
  room_name   TEXT        NOT NULL,
  capacity    INT         NOT NULL DEFAULT 30,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Log aktivitas peserta — basis audit anti-cheat
CREATE TABLE cbt_participant_events (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  participant_id  UUID        NOT NULL REFERENCES cbt_exam_participants(id) ON DELETE CASCADE,
  event_type      TEXT        NOT NULL,
  event_data      JSONB,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sessions: tambah event_id (nullable), class_id menjadi nullable
ALTER TABLE cbt_exam_sessions
  ADD COLUMN event_id UUID REFERENCES cbt_exam_events(id) ON DELETE SET NULL,
  ALTER COLUMN class_id DROP NOT NULL;

-- Participants: tambah kolom anti-cheat + ruangan
ALTER TABLE cbt_exam_participants
  ADD COLUMN room_id            UUID        REFERENCES cbt_exam_rooms(id) ON DELETE SET NULL,
  ADD COLUMN device_fingerprint TEXT,
  ADD COLUMN question_order     JSONB,
  ADD COLUMN last_heartbeat     TIMESTAMPTZ,
  ADD COLUMN app_switch_count   INT         NOT NULL DEFAULT 0,
  ADD COLUMN screenshot_attempt INT         NOT NULL DEFAULT 0,
  ADD COLUMN login_ip           TEXT,
  ADD COLUMN suspicious_flag    BOOLEAN     NOT NULL DEFAULT FALSE;

CREATE INDEX idx_cbt_events_status         ON cbt_exam_events (status);
CREATE INDEX idx_cbt_events_type           ON cbt_exam_events (exam_type, status);
CREATE INDEX idx_cbt_rooms_session         ON cbt_exam_rooms (session_id);
CREATE INDEX idx_cbt_participants_room     ON cbt_exam_participants (room_id);
CREATE INDEX idx_cbt_participants_token    ON cbt_exam_participants (token) WHERE token <> '';
CREATE INDEX idx_participant_events_pid    ON cbt_participant_events (participant_id, created_at DESC);
CREATE INDEX idx_participant_events_type   ON cbt_participant_events (event_type, created_at DESC);
CREATE UNIQUE INDEX idx_cbt_rooms_name     ON cbt_exam_rooms (session_id, room_name);
