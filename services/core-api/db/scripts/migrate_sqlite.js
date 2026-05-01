/**
 * Migrasi data dari SQLite (pusaka.sqlite) ke PostgreSQL.
 *
 * Jalankan:
 *   cd services/core-api/db/scripts
 *   npm install
 *   SQLITE_PATH=../../../apps/web-admin/data/pusaka.sqlite \
 *   DATABASE_URL=postgresql://pusaka:pass@localhost:5432/pusaka \
 *   node migrate_sqlite.js
 */

import Database from 'better-sqlite3';
import pg       from 'pg';
import { randomUUID } from 'crypto';
import path     from 'path';
import { fileURLToPath } from 'url';

const { Pool } = pg;

const __dir     = path.dirname(fileURLToPath(import.meta.url));
const SQLITE_PATH = process.env.SQLITE_PATH
  ?? path.resolve(__dir, '../../../apps/web-admin/data/pusaka.sqlite');
const DATABASE_URL = process.env.DATABASE_URL
  ?? 'postgresql://pusaka:pusaka_dev@localhost:5432/pusaka';

// ── Helpers ───────────────────────────────────────────────────────────────────

/**
 * Normalkan ID ke format UUID (8-4-4-4-12).
 * - UUID standar  → pakai apa adanya
 * - Hex 32 char   → tambah tanda hubung
 * - Format lain (misal "sched-morning") → generate UUID baru
 */
function toUUID(id) {
  if (!id) return randomUUID();
  if (/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(id))
    return id.toLowerCase();
  if (/^[0-9a-f]{32}$/i.test(id)) {
    const h = id.toLowerCase();
    return `${h.slice(0,8)}-${h.slice(8,12)}-${h.slice(12,16)}-${h.slice(16,20)}-${h.slice(20)}`;
  }
  return randomUUID();
}

/** Konversi nilai datetime SQLite ke Date. Nilai null/string literal → now(). */
function toDate(val) {
  if (!val || val === 'CURRENT_TIMESTAMP') return new Date();
  const d = new Date(val);
  return isNaN(d.getTime()) ? new Date() : d;
}

// ── Main ──────────────────────────────────────────────────────────────────────

async function migrate() {
  const sqlite = new Database(SQLITE_PATH, { readonly: true });
  const pool   = new Pool({ connectionString: DATABASE_URL });
  const client = await pool.connect();

  console.log(`SQLite : ${SQLITE_PATH}`);
  console.log(`Postgres: ${DATABASE_URL.replace(/:([^@]+)@/, ':***@')}`);
  console.log('');

  // Peta old_id → new_uuid (dibutuhkan saat insert tabel anak)
  const empMap = new Map();  // employee old_id → uuid
  const jobMap = new Map();  // job old_id → uuid

  try {
    await client.query('BEGIN');

    // ── 1. employees ────────────────────────────────────────────────────────
    const employees = sqlite.prepare('SELECT * FROM employees ORDER BY rowid').all();
    let empOk = 0, empSkip = 0;

    for (const e of employees) {
      const newId = toUUID(e.id);
      empMap.set(e.id, newId);
      try {
        await client.query(
          `INSERT INTO employees
             (id, nip, nama, unit_kerja, is_active, created_at, updated_at)
           VALUES ($1,$2,$3,$4,$5,$6,$7)
           ON CONFLICT (id) DO NOTHING`,
          [newId, e.nip, e.nama, e.unit_kerja,
           e.is_active === 1,
           toDate(e.created_at), toDate(e.updated_at)]
        );
        if (e.pusaka_username || e.pusaka_password) {
          await client.query(
            `INSERT INTO pusaka_accounts
               (employee_id, pusaka_username, pusaka_password, is_enabled, created_at, updated_at)
             VALUES ($1,$2,$3,$4,$5,$6)
             ON CONFLICT (employee_id) DO UPDATE
               SET pusaka_username = EXCLUDED.pusaka_username,
                   pusaka_password = EXCLUDED.pusaka_password,
                   is_enabled      = EXCLUDED.is_enabled,
                   updated_at      = EXCLUDED.updated_at`,
            [newId, e.pusaka_username ?? '', e.pusaka_password ?? '',
             e.is_active === 1,
             toDate(e.created_at), toDate(e.updated_at)]
          );
        }
        empOk++;
      } catch (err) {
        console.warn(`  skip employee ${e.nip}: ${err.message}`);
        empSkip++;
      }
    }
    console.log(`✓ employees      : ${empOk} inserted, ${empSkip} skipped`);

    // ── 2. schedules ────────────────────────────────────────────────────────
    const schedules = sqlite.prepare('SELECT * FROM schedules ORDER BY rowid').all();
    let schedOk = 0, schedSkip = 0;

    for (const s of schedules) {
      const newId = toUUID(s.id);
      try {
        await client.query(
          `INSERT INTO schedules
             (id, label, run_time, run_type, is_enabled, created_at, updated_at)
           VALUES ($1,$2,$3,$4,$5,$6,$7)
           ON CONFLICT (run_type, run_time) DO UPDATE
             SET label      = EXCLUDED.label,
                 run_time   = EXCLUDED.run_time,
                 is_enabled = EXCLUDED.is_enabled,
                 updated_at = EXCLUDED.updated_at`,
          [newId, s.label, s.run_time, s.run_type,
           s.is_enabled === 1,
           toDate(s.created_at), toDate(s.updated_at)]
        );
        schedOk++;
      } catch (err) {
        console.warn(`  skip schedule ${s.id}: ${err.message}`);
        schedSkip++;
      }
    }
    console.log(`✓ schedules      : ${schedOk} upserted, ${schedSkip} skipped`);

    // ── 3. app_settings ─────────────────────────────────────────────────────
    const settings = sqlite.prepare('SELECT * FROM app_settings ORDER BY rowid').all();
    let setOk = 0;

    for (const s of settings) {
      await client.query(
        `INSERT INTO app_settings (key, value, updated_at)
         VALUES ($1,$2,$3)
         ON CONFLICT (key) DO UPDATE
           SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
        [s.key, s.value, toDate(s.updated_at)]
      );
      setOk++;
    }
    console.log(`✓ app_settings   : ${setOk} upserted`);

    // ── 4. jobs ─────────────────────────────────────────────────────────────
    const jobs = sqlite.prepare('SELECT * FROM jobs ORDER BY created_at').all();
    let jobOk = 0, jobSkip = 0;

    for (const j of jobs) {
      const newId    = toUUID(j.id);
      const newEmpId = empMap.get(j.employee_id);
      if (!newEmpId) { jobSkip++; continue; }

      jobMap.set(j.id, newId);
      try {
        await client.query(
          `INSERT INTO jobs
             (id, employee_id, run_type, status, error_message,
              claimed_by, claimed_at, attempts, max_attempts,
              next_retry_at, created_at, updated_at)
           VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
           ON CONFLICT (id) DO NOTHING`,
          [newId, newEmpId, j.run_type, j.status,
           j.error_message ?? '', j.claimed_by ?? '',
           j.claimed_at    ? new Date(j.claimed_at)    : null,
           j.attempts ?? 0, j.max_attempts ?? 3,
           j.next_retry_at ? new Date(j.next_retry_at) : null,
           toDate(j.created_at), toDate(j.updated_at)]
        );
        jobOk++;
      } catch (err) {
        console.warn(`  skip job ${j.id}: ${err.message}`);
        jobSkip++;
      }
    }
    console.log(`✓ jobs           : ${jobOk} inserted, ${jobSkip} skipped`);

    // ── 5. attendance_records ───────────────────────────────────────────────
    const records = sqlite.prepare(
      'SELECT * FROM attendance_records ORDER BY tanggal, created_at'
    ).all();
    let recOk = 0, recSkip = 0;

    for (const r of records) {
      const newId    = toUUID(r.id);
      const newEmpId = empMap.get(r.employee_id);
      const newJobId = jobMap.get(r.source_job_id);

      if (!newEmpId || !newJobId) { recSkip++; continue; }

      try {
        await client.query(
          `INSERT INTO attendance_records
             (id, employee_id, tanggal, jam_masuk, jam_pulang,
              source_job_id, created_at, updated_at)
           VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
           ON CONFLICT (employee_id, tanggal) DO NOTHING`,
          [newId, newEmpId,
           r.tanggal,                  // "YYYY-MM-DD" → PostgreSQL DATE
           r.jam_masuk  ?? '',
           r.jam_pulang ?? '',
           newJobId,
           toDate(r.created_at), toDate(r.updated_at)]
        );
        recOk++;
      } catch (err) {
        console.warn(`  skip record ${r.id}: ${err.message}`);
        recSkip++;
      }
    }
    console.log(`✓ attendance     : ${recOk} inserted, ${recSkip} skipped`);

    await client.query('COMMIT');
    console.log('\n✅ Migrasi selesai!');

  } catch (err) {
    await client.query('ROLLBACK');
    console.error('\n❌ Migrasi gagal (rollback):', err.message);
    process.exit(1);
  } finally {
    client.release();
    await pool.end();
    sqlite.close();
  }
}

migrate();
