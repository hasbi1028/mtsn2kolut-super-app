/**
 * Apply PostgreSQL migrations in backend/db/migrations sequentially.
 *
 * Jalankan:
 *   cd backend/db/scripts
 *   npm install
 *   DATABASE_URL=postgresql://... node apply_migrations.js
 *
 * Default: every migration runs inside one transaction.
 * Online index migration exception:
 *   -- mtsn2kolut:migration non-transactional
 *   CREATE INDEX CONCURRENTLY ...
 * The marker is intentionally narrow and only permits CREATE/DROP INDEX
 * CONCURRENTLY statements executed one by one.
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
const NON_TRANSACTIONAL_MARKER = /^\s*--\s*mtsn2kolut:migration\s+non-transactional\s*$/im;
const CONCURRENT_INDEX_RE = /\b(?:CREATE|DROP)\s+INDEX\s+CONCURRENTLY\b/i;
const TRANSACTION_CONTROL_RE = /^(?:BEGIN|COMMIT|ROLLBACK|SAVEPOINT|RELEASE\s+SAVEPOINT)\b/i;
const NON_TRANSACTIONAL_ALLOWED_RE = /^(?:CREATE|DROP)\s+INDEX\s+CONCURRENTLY\b/i;

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

      const execution = migrationExecutionPlan(file, sql);
      if (execution.mode === 'nontransactional') {
        await applyNonTransactionalMigration(client, file, execution.statements, version, checksum);
      } else {
        await applyTransactionalMigration(client, file, sql, version, checksum);
      }
      baselinePending = false;
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

async function applyTransactionalMigration(client, file, sql, version, checksum) {
  await client.query('BEGIN');
  try {
    await client.query(sql);
    await client.query(
      'INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)',
      [version, checksum],
    );
    await client.query('COMMIT');
    console.log(`applied ${file}`);
  } catch (err) {
    await client.query('ROLLBACK');
    throw new Error(`${file}: ${err.message}`);
  }
}

async function applyNonTransactionalMigration(client, file, statements, version, checksum) {
  try {
    for (const statement of statements) {
      await client.query(statement);
    }
    await client.query(
      'INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)',
      [version, checksum],
    );
    console.log(`applied ${file} (non-transactional)`);
  } catch (err) {
    throw new Error(`${file}: ${err.message}`);
  }
}

export function migrationExecutionPlan(file, sql) {
  const nonTransactional = NON_TRANSACTIONAL_MARKER.test(sql);
  const hasConcurrentIndex = containsConcurrentIndex(sql);

  if (!nonTransactional) {
    if (hasConcurrentIndex) {
      throw new Error(`${file}: CREATE/DROP INDEX CONCURRENTLY requires "-- mtsn2kolut:migration non-transactional"`);
    }
    return { mode: 'transactional', statements: [sql] };
  }

  if (!hasConcurrentIndex) {
    throw new Error(`${file}: non-transactional marker is reserved for online CREATE/DROP INDEX CONCURRENTLY migrations`);
  }

  const statements = splitSqlStatements(sql);
  if (statements.length === 0) {
    throw new Error(`${file}: non-transactional migration has no executable SQL`);
  }
  for (const statement of statements) {
    const executable = stripSqlComments(statement).trim();
    if (TRANSACTION_CONTROL_RE.test(executable)) {
      throw new Error(`${file}: transaction control is not allowed in non-transactional migrations`);
    }
    if (!NON_TRANSACTIONAL_ALLOWED_RE.test(executable)) {
      throw new Error(`${file}: non-transactional migrations may only run CREATE/DROP INDEX CONCURRENTLY statements`);
    }
  }
  return { mode: 'nontransactional', statements };
}

export function containsConcurrentIndex(sql) {
  return CONCURRENT_INDEX_RE.test(stripSqlComments(sql));
}

export function splitSqlStatements(sql) {
  const statements = [];
  let start = 0;
  let i = 0;
  let lineComment = false;
  let blockCommentDepth = 0;
  let singleQuoted = false;
  let doubleQuoted = false;
  let dollarTag = '';

  while (i < sql.length) {
    const ch = sql[i];
    const next = sql[i + 1] ?? '';

    if (lineComment) {
      if (ch === '\n') {
        lineComment = false;
      }
      i += 1;
      continue;
    }
    if (blockCommentDepth > 0) {
      if (ch === '/' && next === '*') {
        blockCommentDepth += 1;
        i += 2;
        continue;
      }
      if (ch === '*' && next === '/') {
        blockCommentDepth -= 1;
        i += 2;
        continue;
      }
      i += 1;
      continue;
    }
    if (dollarTag) {
      if (sql.startsWith(dollarTag, i)) {
        i += dollarTag.length;
        dollarTag = '';
        continue;
      }
      i += 1;
      continue;
    }
    if (singleQuoted) {
      if (ch === "'" && next === "'") {
        i += 2;
        continue;
      }
      if (ch === '\\') {
        i += 2;
        continue;
      }
      if (ch === "'") {
        singleQuoted = false;
      }
      i += 1;
      continue;
    }
    if (doubleQuoted) {
      if (ch === '"' && next === '"') {
        i += 2;
        continue;
      }
      if (ch === '"') {
        doubleQuoted = false;
      }
      i += 1;
      continue;
    }

    if (ch === '-' && next === '-') {
      lineComment = true;
      i += 2;
      continue;
    }
    if (ch === '/' && next === '*') {
      blockCommentDepth = 1;
      i += 2;
      continue;
    }
    if (ch === "'") {
      singleQuoted = true;
      i += 1;
      continue;
    }
    if (ch === '"') {
      doubleQuoted = true;
      i += 1;
      continue;
    }
    if (ch === '$') {
      const tag = readDollarTag(sql, i);
      if (tag) {
        dollarTag = tag;
        i += tag.length;
        continue;
      }
    }
    if (ch === ';') {
      const statement = sql.slice(start, i + 1).trim();
      if (hasExecutableSql(statement)) {
        statements.push(statement);
      }
      start = i + 1;
    }
    i += 1;
  }

  const tail = sql.slice(start).trim();
  if (hasExecutableSql(tail)) {
    statements.push(tail);
  }
  return statements;
}

export function stripSqlComments(sql) {
  let output = '';
  let i = 0;
  let lineComment = false;
  let blockCommentDepth = 0;
  let singleQuoted = false;
  let doubleQuoted = false;
  let dollarTag = '';

  while (i < sql.length) {
    const ch = sql[i];
    const next = sql[i + 1] ?? '';

    if (lineComment) {
      if (ch === '\n') {
        lineComment = false;
        output += '\n';
      }
      i += 1;
      continue;
    }
    if (blockCommentDepth > 0) {
      if (ch === '/' && next === '*') {
        blockCommentDepth += 1;
        i += 2;
        continue;
      }
      if (ch === '*' && next === '/') {
        blockCommentDepth -= 1;
        i += 2;
        continue;
      }
      i += 1;
      continue;
    }
    if (dollarTag) {
      if (sql.startsWith(dollarTag, i)) {
        output += dollarTag;
        i += dollarTag.length;
        dollarTag = '';
        continue;
      }
      output += ch;
      i += 1;
      continue;
    }
    if (singleQuoted) {
      output += ch;
      if (ch === "'" && next === "'") {
        output += next;
        i += 2;
        continue;
      }
      if (ch === '\\') {
        output += next;
        i += 2;
        continue;
      }
      if (ch === "'") {
        singleQuoted = false;
      }
      i += 1;
      continue;
    }
    if (doubleQuoted) {
      output += ch;
      if (ch === '"' && next === '"') {
        output += next;
        i += 2;
        continue;
      }
      if (ch === '"') {
        doubleQuoted = false;
      }
      i += 1;
      continue;
    }

    if (ch === '-' && next === '-') {
      lineComment = true;
      i += 2;
      continue;
    }
    if (ch === '/' && next === '*') {
      blockCommentDepth = 1;
      i += 2;
      continue;
    }
    if (ch === "'") {
      singleQuoted = true;
      output += ch;
      i += 1;
      continue;
    }
    if (ch === '"') {
      doubleQuoted = true;
      output += ch;
      i += 1;
      continue;
    }
    if (ch === '$') {
      const tag = readDollarTag(sql, i);
      if (tag) {
        dollarTag = tag;
        output += tag;
        i += tag.length;
        continue;
      }
    }
    output += ch;
    i += 1;
  }
  return output;
}

function hasExecutableSql(sql) {
  return stripSqlComments(sql).trim() !== '';
}

function readDollarTag(sql, index) {
  const match = sql.slice(index).match(/^\$[A-Za-z_][A-Za-z0-9_]*\$|^\$\$/);
  return match?.[0] ?? '';
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
