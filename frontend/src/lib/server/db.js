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

  // Migrate constraint run_type di schedules & jobs — tambah checkin & checkout
  migrateRunTypeConstraint();

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
  stmt.run('sched-checkin', 'Absen Masuk', '06:30', 'checkin');
  stmt.run('sched-checkout', 'Absen Pulang', '15:01', 'checkout');
} else {
  // Tambah checkin/checkout jika belum ada (data existing mungkin belum punya)
  const existingTypes = db.prepare('SELECT DISTINCT run_type FROM schedules').all().map(r => r.run_type);
  if (!existingTypes.includes('checkin')) {
    db.prepare('INSERT OR IGNORE INTO schedules (id, label, run_time, run_type, is_enabled) VALUES (?, ?, ?, ?, 1)')
      .run('sched-checkin', 'Absen Masuk', '06:30', 'checkin');
  }
  if (!existingTypes.includes('checkout')) {
    db.prepare('INSERT OR IGNORE INTO schedules (id, label, run_time, run_type, is_enabled) VALUES (?, ?, ?, ?, 1)')
      .run('sched-checkout', 'Absen Pulang', '15:01', 'checkout');
  }
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

/**
 * Helper: ambil CREATE TABLE SQL dari sqlite_master
 */
function getCreateSql(name) {
  const row = db.prepare(`SELECT sql FROM sqlite_master WHERE type='table' AND name=?`).get(name);
  return row ? row.sql : '';
}

/**
 * Helper: cek apakah CHECK constraint sudah mencakup semua values
 */
function constraintIncludes(sql, values) {
  const m = sql.match(/CHECK\s*\(\s*run_type\s+IN\s*\(([^)]+)\)\s*\)/i);
  if (!m) return false;
  const existing = m[1].split(',').map((s) => s.trim().replace(/'/g, ''));
  return values.every((v) => existing.includes(v));
}

/**
 * Recreate table dengan expanded CHECK constraint (SQLite tidak support ALTER CONSTRAINT)
 */
function migrateTableConstraint(name, values) {
  const sql = getCreateSql(name);
  const needs = values.map((v) => `'${v}'`).join(',');
  const newConstraint = `CHECK(run_type IN (${needs}))`;

  if (!sql) return;                       // table belum ada
  if (constraintIncludes(sql, values)) return; // sudah up-to-date

  const tableInfo = db.prepare(`PRAGMA table_info(${name})`).all();
  const colDefs = tableInfo.map((c) => {
    let def = `"${c.name}" ${c.type}`;
    if (c.notnull) def += ' NOT NULL';
    if (c.dflt_value !== null) {
      const dv = c.dflt_value.startsWith("'") ? c.dflt_value : `'${c.dflt_value}'`;
      def += ` DEFAULT ${dv}`;
    }
    if (c.pk) def += ' PRIMARY KEY';
    return def;
  });

  const foreignKeys = name === 'jobs'
    ? `, FOREIGN KEY(employee_id) REFERENCES employees(id)`
    : '';

  const indexes = db.prepare(`SELECT sql FROM sqlite_master WHERE type='index' AND tbl_name=? AND sql IS NOT NULL`).all(name);

  // Nonaktifkan FK sementara agar bisa DROP dan recreate
  db.exec('PRAGMA foreign_keys = OFF');

  try {
    // Pass 1: recreate tanpa constraint
    db.exec(`ALTER TABLE ${name} RENAME TO ${name}_old`);
    db.exec(`CREATE TABLE ${name} (${colDefs.join(', ')}${foreignKeys})`);
    db.exec(`INSERT INTO ${name} SELECT * FROM ${name}_old`);
    db.exec(`DROP TABLE ${name}_old`);

    // Pass 2: recreate dengan constraint baru
    db.exec(`ALTER TABLE ${name} RENAME TO ${name}_old`);
    db.exec(`CREATE TABLE ${name} (${colDefs.join(', ')}${foreignKeys}, ${newConstraint})`);
    db.exec(`INSERT INTO ${name} SELECT * FROM ${name}_old`);
    db.exec(`DROP TABLE ${name}_old`);
  } finally {
    db.exec('PRAGMA foreign_keys = ON');
  }

  // Recreate indexes

  // Recreate indexes
  for (const idx of indexes) {
    try { db.exec(idx.sql); } catch { /* ignore if already exists */ }
  }
}

function migrateRunTypeConstraint() {
  ensureJobColumns();
  const newValues = ['morning', 'afternoon', 'checkin', 'checkout'];
  migrateTableConstraint('schedules', newValues);
  migrateTableConstraint('jobs', newValues);
}
