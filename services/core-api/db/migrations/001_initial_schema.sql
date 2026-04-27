-- Migration: 001_initial_schema
-- Database: pusaka
-- Dijalankan otomatis saat docker-compose up pertama kali (initdb.d)
-- atau manual: psql $DATABASE_URL -f 001_initial_schema.sql

-- ── Extensions ────────────────────────────────────────────────────────────────

CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_uuid()

-- ── Enums ─────────────────────────────────────────────────────────────────────

CREATE TYPE run_type_enum AS ENUM (
  'morning',
  'afternoon',
  'checkin',
  'checkout'
);

CREATE TYPE job_status_enum AS ENUM (
  'queued',
  'running',
  'success',
  'failed'
);

-- ── Tables ────────────────────────────────────────────────────────────────────

CREATE TABLE employees (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  nip              TEXT        NOT NULL,
  nama             TEXT        NOT NULL,
  unit_kerja       TEXT        NOT NULL DEFAULT '',
  pusaka_username  TEXT        NOT NULL,
  pusaka_password  TEXT        NOT NULL,
  is_active        BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (nip)
);

CREATE TABLE schedules (
  id         UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
  label      TEXT             NOT NULL,
  run_time   TEXT             NOT NULL,          -- format "HH:MM"
  run_type   run_type_enum    NOT NULL,
  is_enabled BOOLEAN          NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
  UNIQUE (run_type)                              -- satu jadwal per tipe
);

CREATE TABLE jobs (
  id            UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id   UUID              NOT NULL REFERENCES employees(id),
  run_type      run_type_enum     NOT NULL,
  status        job_status_enum   NOT NULL DEFAULT 'queued',
  error_message TEXT              NOT NULL DEFAULT '',
  claimed_by    TEXT              NOT NULL DEFAULT '',
  claimed_at    TIMESTAMPTZ,
  attempts      INTEGER           NOT NULL DEFAULT 0,
  max_attempts  INTEGER           NOT NULL DEFAULT 3,
  next_retry_at TIMESTAMPTZ,
  created_at    TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ       NOT NULL DEFAULT NOW()
);

CREATE TABLE attendance_records (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id   UUID        NOT NULL REFERENCES employees(id),
  tanggal       DATE        NOT NULL,
  jam_masuk     TEXT        NOT NULL DEFAULT '',   -- "HH:MM:SS WITA" atau kosong
  jam_pulang    TEXT        NOT NULL DEFAULT '',
  source_job_id UUID        NOT NULL REFERENCES jobs(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (employee_id, tanggal)
);

CREATE TABLE app_settings (
  key        TEXT        PRIMARY KEY,
  value      TEXT        NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Indexes ───────────────────────────────────────────────────────────────────

CREATE INDEX idx_jobs_status_created_at   ON jobs (status, created_at);
CREATE INDEX idx_jobs_employee_run_status ON jobs (employee_id, run_type, status);
CREATE UNIQUE INDEX uq_jobs_active_employee_run_type
  ON jobs (employee_id, run_type)
  WHERE status IN ('queued', 'running');
CREATE INDEX idx_jobs_next_retry_at       ON jobs (next_retry_at) WHERE next_retry_at IS NOT NULL;
CREATE INDEX idx_attendance_tanggal       ON attendance_records (tanggal DESC);
CREATE INDEX idx_attendance_employee      ON attendance_records (employee_id, tanggal DESC);

-- ── Seed: jadwal default ──────────────────────────────────────────────────────
-- Di-seed di sini agar setiap install fresh langsung siap pakai.
-- Script migrasi dari SQLite akan UPSERT di atas ini.

INSERT INTO schedules (id, label, run_time, run_type, is_enabled) VALUES
  (gen_random_uuid(), 'Pagi',         '07:00', 'morning',   TRUE),
  (gen_random_uuid(), 'Sore',         '16:00', 'afternoon', TRUE),
  (gen_random_uuid(), 'Absen Masuk',  '06:30', 'checkin',   FALSE),
  (gen_random_uuid(), 'Absen Pulang', '15:01', 'checkout',  FALSE)
ON CONFLICT (run_type) DO NOTHING;
