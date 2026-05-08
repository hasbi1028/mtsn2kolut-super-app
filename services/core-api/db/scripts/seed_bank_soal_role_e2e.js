#!/usr/bin/env node

/**
 * Idempotent seed for Bank Soal browser role smoke accounts.
 *
 * Usage:
 *   DATABASE_URL=postgresql://... npm run seed:bank-soal-e2e-roles
 *
 * This is a manual/scripted seed, not a migration. It creates controlled E2E
 * users and custom RBAC roles for browser smoke checks without storing any real
 * credential in the repository.
 */

import pg from 'pg';
import { pathToFileURL } from 'node:url';

const { Pool } = pg;

const DEFAULT_USERNAME_PREFIX = 'e2e_bank_soal_';

export const seedPersonas = [
  {
    key: 'admin',
    envPrefix: 'ADMIN',
    displayName: 'E2E Bank Soal Admin',
    roleCode: 'admin',
    roleName: 'Administrator',
    roleDescription: 'Existing system administrator role.',
    manageRolePermissions: false,
    permissions: [],
  },
  {
    key: 'creator',
    envPrefix: 'CREATOR',
    displayName: 'E2E Bank Soal Creator',
    roleCode: 'bank_soal_e2e_creator',
    roleName: 'Bank Soal E2E Creator',
    roleDescription: 'Controlled E2E role for Bank Soal authoring smoke.',
    manageRolePermissions: true,
    permissions: ['bank_soal.analytics', 'bank_soal.create', 'bank_soal.read', 'bank_soal.update'],
  },
  {
    key: 'reviewer',
    envPrefix: 'REVIEWER',
    displayName: 'E2E Bank Soal Reviewer',
    roleCode: 'bank_soal_e2e_reviewer',
    roleName: 'Bank Soal E2E Reviewer',
    roleDescription: 'Controlled E2E role for Bank Soal verification smoke.',
    manageRolePermissions: true,
    permissions: ['bank_soal.analytics', 'bank_soal.read', 'bank_soal.review'],
  },
  {
    key: 'importer',
    envPrefix: 'IMPORTER',
    displayName: 'E2E Bank Soal Importer',
    roleCode: 'bank_soal_e2e_importer',
    roleName: 'Bank Soal E2E Importer',
    roleDescription: 'Controlled E2E role for Bank Soal import smoke.',
    manageRolePermissions: true,
    permissions: ['bank_soal.analytics', 'bank_soal.import', 'bank_soal.read'],
  },
  {
    key: 'readonly',
    envPrefix: 'READONLY',
    displayName: 'E2E Bank Soal Readonly',
    roleCode: 'bank_soal_e2e_readonly',
    roleName: 'Bank Soal E2E Readonly',
    roleDescription: 'Controlled E2E role for read-only Bank Soal smoke.',
    manageRolePermissions: true,
    permissions: ['bank_soal.analytics', 'bank_soal.read'],
  },
  {
    key: 'no-access',
    envPrefix: 'NO_ACCESS',
    displayName: 'E2E Bank Soal No Access',
    roleCode: 'bank_soal_e2e_no_access',
    roleName: 'Bank Soal E2E No Access',
    roleDescription: 'Controlled E2E role with no Bank Soal permission.',
    manageRolePermissions: true,
    permissions: ['settings.account'],
  },
];

export function resolveDatabaseURL(env = process.env) {
  if (env.DATABASE_URL) return env.DATABASE_URL;
  if (truthy(env.ALLOW_LOCAL_DATABASE_URL) && env.LOCAL_DATABASE_URL) return env.LOCAL_DATABASE_URL;
  throw new Error('DATABASE_URL is required; set ALLOW_LOCAL_DATABASE_URL=true with LOCAL_DATABASE_URL only for explicit local development');
}

export function truthy(value) {
  return ['1', 'true', 'yes'].includes(String(value ?? '').toLowerCase());
}

export function personaSmokeEnv(personaOrKey) {
  const persona = findPersona(personaOrKey);
  return {
    username: `WEB_ADMIN_BANK_SOAL_E2E_${persona.envPrefix}_USERNAME`,
    password: `WEB_ADMIN_BANK_SOAL_E2E_${persona.envPrefix}_PASSWORD`,
  };
}

export function personaSeedEnv(personaOrKey) {
  const persona = findPersona(personaOrKey);
  return {
    username: `BANK_SOAL_E2E_${persona.envPrefix}_USERNAME`,
    password: `BANK_SOAL_E2E_${persona.envPrefix}_PASSWORD`,
  };
}

export function personaCredentials(personaOrKey, env = process.env) {
  const persona = findPersona(personaOrKey);
  const seedEnv = personaSeedEnv(persona);
  const usernamePrefix = env.BANK_SOAL_E2E_USERNAME_PREFIX ?? DEFAULT_USERNAME_PREFIX;
  return {
    username: env[seedEnv.username] ?? `${usernamePrefix}${persona.key.replaceAll('-', '_')}`,
    password: env[seedEnv.password] ?? env.BANK_SOAL_E2E_PASSWORD,
  };
}

export function rolePermissionPlan() {
  return Object.fromEntries(
    seedPersonas.map((persona) => [
      persona.key,
      {
        roleCode: persona.roleCode,
        manageRolePermissions: persona.manageRolePermissions,
        permissions: [...persona.permissions].sort(),
      },
    ]),
  );
}

export function smokeEnvTemplate(env = process.env) {
  const lines = [];
  for (const persona of seedPersonas) {
    const smokeEnv = personaSmokeEnv(persona);
    const credentials = personaCredentials(persona, env);
    lines.push(`${smokeEnv.username}=${credentials.username}`);
    lines.push(`${smokeEnv.password}=<set BANK_SOAL_E2E_PASSWORD or ${personaSeedEnv(persona).password}>`);
  }
  return lines.join('\n');
}

export async function seedBankSoalRoleE2E(options = {}) {
  const env = options.env ?? process.env;
  const databaseURL = options.databaseURL ?? resolveDatabaseURL(env);
  validateUniqueUsernames(env);
  validateRequiredPasswords(env);
  const pool = options.pool ?? new Pool({ connectionString: databaseURL });
  const client = await pool.connect();
  const shouldEndPool = !options.pool;

  try {
    await client.query('BEGIN');
    await ensurePgcrypto(client);
    await ensureRequiredPermissions(client);
    await ensureRoles(client);

    const results = [];
    for (const persona of seedPersonas) {
      const credentials = personaCredentials(persona, env);
      const user = await upsertPersonaUser(client, persona, credentials);
      await assignPersonaRole(client, user.id, persona);
      results.push({
        persona: persona.key,
        username: credentials.username,
        roleCode: persona.roleCode,
        existed: user.existed,
        passwordChanged: user.password_changed,
        sessionsRevoked: user.sessions_revoked,
      });
    }

    await client.query('COMMIT');
    return results;
  } catch (err) {
    await client.query('ROLLBACK').catch(() => undefined);
    throw err;
  } finally {
    client.release();
    if (shouldEndPool) await pool.end();
  }
}

async function ensurePgcrypto(client) {
  const result = await client.query(`
    SELECT
      to_regprocedure('crypt(text,text)') IS NOT NULL AS has_crypt,
      to_regprocedure('gen_salt(text,integer)') IS NOT NULL AS has_gen_salt
  `);
  if (!result.rows[0]?.has_crypt || !result.rows[0]?.has_gen_salt) {
    throw new Error('pgcrypto crypt/gen_salt functions are unavailable; run migrations first');
  }
}

async function ensureRequiredPermissions(client) {
  const required = [...new Set(seedPersonas.flatMap((persona) => persona.permissions))].sort();
  if (required.length === 0) return;

  const result = await client.query(
    'SELECT code FROM rbac_permissions WHERE code = ANY($1::text[]) AND is_active = TRUE',
    [required],
  );
  const found = new Set(result.rows.map((row) => row.code));
  const missing = required.filter((permission) => !found.has(permission));
  if (missing.length > 0) {
    throw new Error(`missing RBAC permissions; run migrations first: ${missing.join(', ')}`);
  }
}

async function ensureRoles(client) {
  for (const persona of seedPersonas) {
    if (!persona.manageRolePermissions) {
      const role = await client.query('SELECT id FROM rbac_roles WHERE code = $1 AND is_active = TRUE', [persona.roleCode]);
      if (role.rowCount < 1) throw new Error(`missing active RBAC role: ${persona.roleCode}`);
      continue;
    }

    await client.query(
      `
        INSERT INTO rbac_roles (code, name, description, is_system, is_active)
        VALUES ($1, $2, $3, FALSE, TRUE)
        ON CONFLICT (code) DO UPDATE
        SET name = EXCLUDED.name,
            description = EXCLUDED.description,
            is_system = FALSE,
            is_active = TRUE,
            updated_at = NOW()
      `,
      [persona.roleCode, persona.roleName, persona.roleDescription],
    );

    await client.query(
      'DELETE FROM rbac_role_permissions WHERE role_id = (SELECT id FROM rbac_roles WHERE code = $1)',
      [persona.roleCode],
    );

    if (persona.permissions.length > 0) {
      await client.query(
        `
          INSERT INTO rbac_role_permissions (role_id, permission_id)
          SELECT r.id, p.id
          FROM rbac_roles r
          JOIN rbac_permissions p ON p.code = ANY($2::text[])
          WHERE r.code = $1
            AND r.is_active = TRUE
            AND p.is_active = TRUE
          ON CONFLICT (role_id, permission_id) DO NOTHING
        `,
        [persona.roleCode, persona.permissions],
      );
    }
  }
}

async function upsertPersonaUser(client, persona, credentials) {
  if (!credentials.username.trim()) throw new Error(`empty username for persona ${persona.key}`);
  if (!credentials.password.trim()) throw new Error(`empty password for persona ${persona.key}`);

  const result = await client.query(
    `
      WITH existing AS (
        SELECT id, password_hash = crypt($2, password_hash) AS password_matches
        FROM users
        WHERE username = $1
      ),
      upserted AS (
        INSERT INTO users (
          username,
          password_hash,
          display_name,
          is_active,
          must_change_password,
          password_changed_at,
          deleted_at
        )
        VALUES ($1, crypt($2, gen_salt('bf', 10)), $3, TRUE, FALSE, NOW(), NULL)
        ON CONFLICT (username) DO UPDATE
        SET display_name = EXCLUDED.display_name,
            password_hash = CASE
              WHEN COALESCE((SELECT password_matches FROM existing), FALSE)
                THEN users.password_hash
              ELSE crypt($2, gen_salt('bf', 10))
            END,
            is_active = TRUE,
            must_change_password = FALSE,
            password_changed_at = CASE
              WHEN COALESCE((SELECT password_matches FROM existing), FALSE)
                THEN users.password_changed_at
              ELSE NOW()
            END,
            deleted_at = NULL,
            auth_version = CASE
              WHEN COALESCE((SELECT password_matches FROM existing), FALSE)
                THEN users.auth_version
              ELSE users.auth_version + 1
            END,
            updated_at = NOW()
        RETURNING id
      ),
      revoked AS (
        UPDATE auth_sessions
        SET revoked_at = COALESCE(revoked_at, NOW()),
            updated_at = NOW()
        WHERE user_id = (SELECT id FROM upserted)
          AND revoked_at IS NULL
          AND NOT COALESCE((SELECT password_matches FROM existing), FALSE)
        RETURNING id
      )
      SELECT
        upserted.id,
        COALESCE((SELECT TRUE FROM existing), FALSE) AS existed,
        NOT COALESCE((SELECT password_matches FROM existing), FALSE) AS password_changed,
        (SELECT COUNT(*)::int FROM revoked) AS sessions_revoked
      FROM upserted
    `,
    [credentials.username, credentials.password, persona.displayName],
  );

  if (result.rowCount < 1) throw new Error(`failed to upsert persona ${persona.key}`);
  return result.rows[0];
}

async function assignPersonaRole(client, userID, persona) {
  await client.query('DELETE FROM rbac_user_roles WHERE user_id = $1', [userID]);
  await client.query(
    `
      INSERT INTO rbac_user_roles (user_id, role_id)
      SELECT $1, r.id
      FROM rbac_roles r
      WHERE r.code = $2
        AND r.is_active = TRUE
      ON CONFLICT (user_id, role_id) DO NOTHING
    `,
    [userID, persona.roleCode],
  );
}

function findPersona(personaOrKey) {
  if (typeof personaOrKey === 'object' && personaOrKey?.key) return personaOrKey;
  const persona = seedPersonas.find((item) => item.key === personaOrKey);
  if (!persona) throw new Error(`unknown Bank Soal E2E persona: ${personaOrKey}`);
  return persona;
}


export function validateRequiredPasswords(env) {
  const missing = [];
  for (const persona of seedPersonas) {
    const seedEnv = personaSeedEnv(persona);
    if (!env[seedEnv.password] && !env.BANK_SOAL_E2E_PASSWORD) missing.push(seedEnv.password);
  }
  if (missing.length > 0) {
    throw new Error(`BANK_SOAL_E2E_PASSWORD or per-persona password env is required: ${missing.join(', ')}`);
  }
}

export function validateUniqueUsernames(env) {
  const seen = new Map();
  for (const persona of seedPersonas) {
    const { username } = personaCredentials(persona, env);
    const existing = seen.get(username);
    if (existing) {
      throw new Error(`duplicate Bank Soal E2E username "${username}" for ${existing} and ${persona.key}`);
    }
    seen.set(username, persona.key);
  }
}

function maskDatabaseURL(value) {
  return value.replace(/:([^:@/]+)@/, ':***@');
}

function printSummary(databaseURL, results) {
  console.log(`Seeded Bank Soal E2E role personas in ${maskDatabaseURL(databaseURL)}`);
  for (const result of results) {
    const action = result.existed ? 'updated' : 'created';
    const password = result.passwordChanged ? 'password set' : 'password unchanged';
    console.log(`- ${result.persona}: ${result.username} (${result.roleCode}, ${action}, ${password})`);
  }
  console.log('');
  console.log('Use these usernames with npm run smoke:bank-soal:roles; keep passwords in env only.');
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  let databaseURL;
  try {
    databaseURL = resolveDatabaseURL(process.env);
  } catch (err) {
    console.error('Bank Soal E2E role seed failed:', err.message);
    process.exit(1);
  }
  seedBankSoalRoleE2E({ databaseURL })
    .then((results) => printSummary(databaseURL, results))
    .catch((err) => {
      console.error('Bank Soal E2E role seed failed:', err.message);
      process.exit(1);
    });
}
