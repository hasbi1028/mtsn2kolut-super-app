import BetterSqlite3 from 'better-sqlite3';
import { drizzle } from 'drizzle-orm/better-sqlite3';
import fs from 'fs';
import path from 'path';
import * as schema from './schema.js';

const DB_PATH = process.env.DB_PATH ?? path.resolve('../data/pusaka.sqlite');
fs.mkdirSync(path.dirname(DB_PATH), { recursive: true });

const sqlite = new BetterSqlite3(DB_PATH);
sqlite.pragma('journal_mode = WAL');
sqlite.pragma('foreign_keys = ON');

// ── Bootstrap tables ─────────────────────────────────────────────────────────
// CREATE TABLE IF NOT EXISTS keeps this idempotent for new installs.
// Drizzle migrations handle schema changes going forward.

sqlite.exec(`
CREATE TABLE IF NOT EXISTS employees (
  id               TEXT PRIMARY KEY,
  nip              TEXT NOT NULL UNIQUE,
  nama             TEXT NOT NULL,
  unit_kerja       TEXT NOT NULL DEFAULT '',
  pusaka_username  TEXT NOT NULL,
  pusaka_password  TEXT NOT NULL,
  is_active        INTEGER NOT NULL DEFAULT 1,
  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS schedules (
  id          TEXT PRIMARY KEY,
  label       TEXT NOT NULL,
  run_time    TEXT NOT NULL,
  run_type    TEXT NOT NULL CHECK(run_type IN ('morning','afternoon','checkin','checkout')),
  is_enabled  INTEGER NOT NULL DEFAULT 1,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS jobs (
  id            TEXT PRIMARY KEY,
  employee_id   TEXT NOT NULL,
  run_type      TEXT NOT NULL CHECK(run_type IN ('morning','afternoon','checkin','checkout')),
  status        TEXT NOT NULL CHECK(status IN ('queued','running','success','failed')),
  error_message TEXT NOT NULL DEFAULT '',
  claimed_by    TEXT NOT NULL DEFAULT '',
  claimed_at    DATETIME,
  attempts      INTEGER NOT NULL DEFAULT 0,
  max_attempts  INTEGER NOT NULL DEFAULT 3,
  next_retry_at DATETIME,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(employee_id) REFERENCES employees(id)
);

CREATE TABLE IF NOT EXISTS attendance_records (
  id            TEXT PRIMARY KEY,
  employee_id   TEXT NOT NULL,
  tanggal       TEXT NOT NULL,
  jam_masuk     TEXT NOT NULL DEFAULT '',
  jam_pulang    TEXT NOT NULL DEFAULT '',
  source_job_id TEXT NOT NULL,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(employee_id, tanggal),
  FOREIGN KEY(employee_id)   REFERENCES employees(id),
  FOREIGN KEY(source_job_id) REFERENCES jobs(id)
);

CREATE TABLE IF NOT EXISTS app_settings (
  key        TEXT PRIMARY KEY,
  value      TEXT NOT NULL DEFAULT '',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_jobs_status_created_at    ON jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_employee_run_status  ON jobs(employee_id, run_type, status);
CREATE INDEX IF NOT EXISTS idx_jobs_next_retry_at        ON jobs(next_retry_at);
`);

migrateSchema();
seedDefaults();

export const db = drizzle(sqlite, { schema });

// raw sqlite instance for complex queries that benefit from direct prepared statements
export const rawDb = sqlite;

// ── Schema migrations ─────────────────────────────────────────────────────────

function migrateSchema() {
	// Ensure jobs table has all columns (older installs may be missing some)
	const cols = sqlite.prepare('PRAGMA table_info(jobs)').all().map((c: { name: string }) => c.name);
	if (!cols.includes('claimed_by'))    sqlite.exec(`ALTER TABLE jobs ADD COLUMN claimed_by TEXT NOT NULL DEFAULT ''`);
	if (!cols.includes('claimed_at'))    sqlite.exec(`ALTER TABLE jobs ADD COLUMN claimed_at DATETIME`);
	if (!cols.includes('attempts'))      sqlite.exec(`ALTER TABLE jobs ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0`);
	if (!cols.includes('max_attempts'))  sqlite.exec(`ALTER TABLE jobs ADD COLUMN max_attempts INTEGER NOT NULL DEFAULT 3`);
	if (!cols.includes('next_retry_at')) sqlite.exec(`ALTER TABLE jobs ADD COLUMN next_retry_at DATETIME`);

	// Migrate CHECK constraint on run_type to include checkin/checkout
	migrateRunTypeConstraint();
}

function migrateRunTypeConstraint() {
	const newValues: string[] = ['morning', 'afternoon', 'checkin', 'checkout'];

	for (const table of ['schedules', 'jobs'] as const) {
		const row = sqlite.prepare(`SELECT sql FROM sqlite_master WHERE type='table' AND name=?`).get(table) as { sql: string } | undefined;
		if (!row) continue;

		const m = row.sql.match(/CHECK\s*\(\s*run_type\s+IN\s*\(([^)]+)\)\s*\)/i);
		const existing = m ? m[1].split(',').map((s: string) => s.trim().replace(/'/g, '')) : [];
		if (newValues.every((v) => existing.includes(v))) continue;

		const tableInfo = sqlite.prepare(`PRAGMA table_info(${table})`).all() as Array<{
			name: string; type: string; notnull: number; dflt_value: string | null; pk: number;
		}>;
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

		const fk = table === 'jobs' ? `, FOREIGN KEY(employee_id) REFERENCES employees(id)` : '';
		const need = newValues.map((v) => `'${v}'`).join(',');
		const indexes = sqlite.prepare(`SELECT sql FROM sqlite_master WHERE type='index' AND tbl_name=? AND sql IS NOT NULL`).all(table) as Array<{ sql: string }>;

		sqlite.pragma('foreign_keys = OFF');
		try {
			sqlite.exec(`
				ALTER TABLE ${table} RENAME TO ${table}_old;
				CREATE TABLE ${table} (${colDefs.join(', ')}${fk}, CHECK(run_type IN (${need})));
				INSERT INTO ${table} SELECT * FROM ${table}_old;
				DROP TABLE ${table}_old;
			`);
		} finally {
			sqlite.pragma('foreign_keys = ON');
		}

		for (const idx of indexes) {
			try { sqlite.exec(idx.sql); } catch { /* ignore duplicate */ }
		}
	}
}

// ── Seed defaults ─────────────────────────────────────────────────────────────

function seedDefaults() {
	const existingTypes = sqlite.prepare('SELECT DISTINCT run_type FROM schedules').all().map((r: { run_type: string }) => r.run_type);
	const defaults = [
		{ id: 'sched-morning',  label: 'Pagi',         run_time: '07:00', run_type: 'morning' },
		{ id: 'sched-afternoon',label: 'Sore',          run_time: '16:00', run_type: 'afternoon' },
		{ id: 'sched-checkin',  label: 'Absen Masuk',   run_time: '06:30', run_type: 'checkin' },
		{ id: 'sched-checkout', label: 'Absen Pulang',  run_time: '15:01', run_type: 'checkout' },
	];
	const stmt = sqlite.prepare('INSERT OR IGNORE INTO schedules (id, label, run_time, run_type, is_enabled) VALUES (?, ?, ?, ?, 1)');
	for (const s of defaults) {
		if (!existingTypes.includes(s.run_type)) stmt.run(s.id, s.label, s.run_time, s.run_type);
	}

	const settingsDefaults: [string, string][] = [['max_concurrent', '1'], ['headless', '0']];
	const settingStmt = sqlite.prepare('INSERT OR IGNORE INTO app_settings (key, value) VALUES (?, ?)');
	for (const [key, value] of settingsDefaults) settingStmt.run(key, value);
}
