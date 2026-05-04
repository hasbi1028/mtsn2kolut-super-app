/**
 * Apply PostgreSQL migrations in backend/db/migrations sequentially.
 *
 * Jalankan:
 *   cd backend/db/scripts
 *   npm install
 *   DATABASE_URL=postgresql://... node apply_migrations.js
 */

import fs from 'fs';
import crypto from 'crypto';
import path from 'path';
import pg from 'pg';
import { fileURLToPath, pathToFileURL } from 'url';

const { Pool } = pg;

const __dir = path.dirname(fileURLToPath(import.meta.url));
const LOCAL_DATABASE_URL = 'postgresql://pusaka:pusaka_dev@localhost:5432/pusaka';
const DEFAULT_MIGRATIONS_DIR = path.resolve(__dir, '../migrations');

export async function hasExistingSchema(client) {
  const res = await client.query(`
    SELECT COUNT(*)::int AS count
    FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_name IN ('employees', 'jobs', 'attendance_records', 'schedules', 'app_settings')
  `);
  return (res.rows[0]?.count ?? 0) > 0;
}

export async function applyMigrations(options = {}) {
  const env = options.env ?? process.env;
  const migrationsDir = options.migrationsDir ?? DEFAULT_MIGRATIONS_DIR;
  const databaseURL = options.databaseURL ?? resolveDatabaseURL(env);
  const baselineOnExistingSchema = options.baselineOnExistingSchema ?? baselineOnExistingSchemaEnabled(env);
  const pool = options.pool ?? new Pool({ connectionString: databaseURL });
  const client = await pool.connect();

  try {
    await client.query(`
      CREATE TABLE IF NOT EXISTS schema_migrations (
        version TEXT PRIMARY KEY,
        checksum TEXT NOT NULL DEFAULT '',
        applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
      )
    `);
    await client.query(`
      ALTER TABLE schema_migrations
      ADD COLUMN IF NOT EXISTS checksum TEXT NOT NULL DEFAULT ''
    `);

    const files = fs.readdirSync(migrationsDir)
      .filter((name) => name.endsWith('.sql'))
      .sort();
    validateMigrationOrder(files);

    let baselinePending = baselineOnExistingSchema && await hasExistingSchema(client);

    for (const file of files) {
      const version = file.replace(/\.sql$/, '');
      const acceptableVersions = [version, ...legacyMigrationVersions(version)];
      const fullPath = path.join(migrationsDir, file);
      const sql = fs.readFileSync(fullPath, 'utf8');
      const checksum = checksumFor(sql);
      const exists = await client.query(
        'SELECT version, checksum FROM schema_migrations WHERE version = ANY($1::text[])',
        [acceptableVersions],
      );
      if (exists.rowCount > 0) {
        const appliedVersion = exists.rows[0]?.version ?? version;
        const appliedChecksum = exists.rows[0]?.checksum ?? '';
        assertAppliedChecksum(file, appliedVersion, appliedChecksum, checksum);
        if (!appliedChecksum) {
          await client.query(
            'UPDATE schema_migrations SET checksum = $2 WHERE version = $1',
            [appliedVersion, checksum],
          );
        }
        console.log(`skip ${file}`);
        continue;
      }

      if (baselinePending && version === '001_initial_schema') {
        await client.query(
          'INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)',
          [version, checksum],
        );
        console.log(`baseline ${file}`);
        baselinePending = false;
        continue;
      }

      await client.query('BEGIN');
      try {
        await client.query(sql);
        await client.query(
          'INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)',
          [version, checksum],
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

export function resolveDatabaseURL(env = process.env) {
  if (env.DATABASE_URL) {
    return env.DATABASE_URL;
  }
  if (truthy(env.ALLOW_LOCAL_DATABASE_URL)) {
    return LOCAL_DATABASE_URL;
  }
  throw new Error('DATABASE_URL is required; set ALLOW_LOCAL_DATABASE_URL=true only for explicit local development');
}

export function baselineOnExistingSchemaEnabled(env = process.env) {
  return !['0', 'false', 'no'].includes(String(env.BASELINE_ON_EXISTING_SCHEMA ?? 'true').toLowerCase());
}

export function truthy(value) {
  return ['1', 'true', 'yes'].includes(String(value ?? '').toLowerCase());
}

export function checksumFor(sql) {
  return crypto.createHash('sha256').update(sql).digest('hex');
}

export function validateMigrationOrder(files) {
  const seenNumericPrefixes = new Map();
  for (const file of files) {
    const match = file.match(/^(\d+)_/);
    if (!match) {
      continue;
    }
    const prefix = match[1];
    const existing = seenNumericPrefixes.get(prefix);
    if (existing) {
      throw new Error(`duplicate migration prefix ${prefix}: ${existing}, ${file}`);
    }
    seenNumericPrefixes.set(prefix, file);
  }
}

export function assertAppliedChecksum(file, appliedVersion, appliedChecksum, checksum) {
  if (appliedChecksum && appliedChecksum !== checksum) {
    throw new Error(`checksum mismatch for ${file} (applied as ${appliedVersion})`);
  }
}

export function legacyMigrationVersions(version) {
  if (version === '027a_library_foundation') {
    return ['027_library_foundation'];
  }
  return [];
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  applyMigrations().catch((err) => {
    console.error('migration failed:', err.message);
    process.exit(1);
  });
}
