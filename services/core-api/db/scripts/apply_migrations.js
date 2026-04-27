/**
 * Apply PostgreSQL migrations in backend/db/migrations sequentially.
 *
 * Jalankan:
 *   cd backend/db/scripts
 *   npm install
 *   DATABASE_URL=postgresql://... node apply_migrations.js
 */

import fs from 'fs';
import path from 'path';
import pg from 'pg';
import { fileURLToPath } from 'url';

const { Pool } = pg;

const __dir = path.dirname(fileURLToPath(import.meta.url));
const MIGRATIONS_DIR = path.resolve(__dir, '../migrations');
const DATABASE_URL = process.env.DATABASE_URL
  ?? 'postgresql://pusaka:pusaka_dev@localhost:5432/pusaka';
const BASELINE_ON_EXISTING_SCHEMA = !['0', 'false', 'no'].includes(
  String(process.env.BASELINE_ON_EXISTING_SCHEMA ?? 'true').toLowerCase()
);

async function hasExistingSchema(client) {
  const res = await client.query(`
    SELECT COUNT(*)::int AS count
    FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_name IN ('employees', 'jobs', 'attendance_records', 'schedules', 'app_settings')
  `);
  return (res.rows[0]?.count ?? 0) > 0;
}

async function applyMigrations() {
  const pool = new Pool({ connectionString: DATABASE_URL });
  const client = await pool.connect();

  try {
    await client.query(`
      CREATE TABLE IF NOT EXISTS schema_migrations (
        version TEXT PRIMARY KEY,
        applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
      )
    `);

    const files = fs.readdirSync(MIGRATIONS_DIR)
      .filter((name) => name.endsWith('.sql'))
      .sort();

    let baselinePending = BASELINE_ON_EXISTING_SCHEMA && await hasExistingSchema(client);

    for (const file of files) {
      const version = file.replace(/\.sql$/, '');
      const exists = await client.query(
        'SELECT 1 FROM schema_migrations WHERE version = $1',
        [version],
      );
      if (exists.rowCount > 0) {
        console.log(`skip ${file}`);
        continue;
      }

      if (baselinePending && version === '001_initial_schema') {
        await client.query(
          'INSERT INTO schema_migrations (version) VALUES ($1)',
          [version],
        );
        console.log(`baseline ${file}`);
        baselinePending = false;
        continue;
      }

      const fullPath = path.join(MIGRATIONS_DIR, file);
      const sql = fs.readFileSync(fullPath, 'utf8');

      await client.query('BEGIN');
      try {
        await client.query(sql);
        await client.query(
          'INSERT INTO schema_migrations (version) VALUES ($1)',
          [version],
        );
        await client.query('COMMIT');
        console.log(`applied ${file}`);
        baselinePending = false;
      } catch (err) {
        await client.query('ROLLBACK');
        throw new Error(`${file}: ${err.message}`);
      }
    }

    console.log('migrations complete');
  } finally {
    client.release();
    await pool.end();
  }
}

applyMigrations().catch((err) => {
  console.error('migration failed:', err.message);
  process.exit(1);
});
