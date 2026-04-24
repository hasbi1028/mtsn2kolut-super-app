import Database from 'better-sqlite3';
import fs from 'fs';
import path from 'path';

const DB_PATH = process.env.DB_PATH || path.resolve('../data/pusaka.sqlite');
fs.mkdirSync(path.dirname(DB_PATH), { recursive: true });

export const db = new Database(DB_PATH);
db.pragma('journal_mode = WAL');

db.exec(`
CREATE TABLE IF NOT EXISTS employees (
  id TEXT PRIMARY KEY,
  nip TEXT NOT NULL UNIQUE,
  nama TEXT NOT NULL,
  unit_kerja TEXT NOT NULL DEFAULT '',
  pusaka_username TEXT NOT NULL,
  pusaka_password TEXT NOT NULL,
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS schedules (
  id TEXT PRIMARY KEY,
  label TEXT NOT NULL,
  run_time TEXT NOT NULL,
  run_type TEXT NOT NULL CHECK(run_type IN ('morning','afternoon')),
  is_enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  employee_id TEXT NOT NULL,
  run_type TEXT NOT NULL,
  status TEXT NOT NULL CHECK(status IN ('queued','running','success','failed')),
  error_message TEXT NOT NULL DEFAULT '',
  claimed_by TEXT NOT NULL DEFAULT '',
  claimed_at DATETIME,
  attempts INTEGER NOT NULL DEFAULT 0,
  max_attempts INTEGER NOT NULL DEFAULT 3,
  next_retry_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(employee_id) REFERENCES employees(id)
);

CREATE TABLE IF NOT EXISTS attendance_records (
  id TEXT PRIMARY KEY,
  employee_id TEXT NOT NULL,
  tanggal TEXT NOT NULL,
  jam_masuk TEXT NOT NULL DEFAULT '',
  jam_pulang TEXT NOT NULL DEFAULT '',
  source_job_id TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(employee_id, tanggal),
  FOREIGN KEY(employee_id) REFERENCES employees(id),
  FOREIGN KEY(source_job_id) REFERENCES jobs(id)
);

CREATE TABLE IF NOT EXISTS app_settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL DEFAULT '',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`);

ensureJobColumns();

db.exec(`
CREATE INDEX IF NOT EXISTS idx_jobs_status_created_at ON jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_employee_run_status ON jobs(employee_id, run_type, status);
CREATE INDEX IF NOT EXISTS idx_jobs_next_retry_at ON jobs(next_retry_at);
`);

const hasDefaultSchedules = db.prepare('SELECT COUNT(*) AS n FROM schedules').get().n > 0;
if (!hasDefaultSchedules) {
  const stmt = db.prepare('INSERT INTO schedules (id, label, run_time, run_type, is_enabled) VALUES (?, ?, ?, ?, 1)');
  stmt.run('sched-morning', 'Pagi', '07:00', 'morning');
  stmt.run('sched-afternoon', 'Sore', '16:00', 'afternoon');
}

const defaultSettings = [
  ['max_concurrent', '1'],
  ['headless', '0']
];
for (const [key, value] of defaultSettings) {
  db.prepare('INSERT OR IGNORE INTO app_settings (key, value) VALUES (?, ?)').run(key, value);
}

function ensureJobColumns() {
  const cols = db.prepare(`PRAGMA table_info(jobs)`).all().map((c) => c.name);

  if (!cols.includes('claimed_by')) db.exec(`ALTER TABLE jobs ADD COLUMN claimed_by TEXT NOT NULL DEFAULT ''`);
  if (!cols.includes('claimed_at')) db.exec(`ALTER TABLE jobs ADD COLUMN claimed_at DATETIME`);
  if (!cols.includes('attempts')) db.exec(`ALTER TABLE jobs ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0`);
  if (!cols.includes('max_attempts')) db.exec(`ALTER TABLE jobs ADD COLUMN max_attempts INTEGER NOT NULL DEFAULT 3`);
  if (!cols.includes('next_retry_at')) db.exec(`ALTER TABLE jobs ADD COLUMN next_retry_at DATETIME`);
}
