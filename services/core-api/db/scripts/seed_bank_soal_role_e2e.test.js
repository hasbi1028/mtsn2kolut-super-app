import assert from 'node:assert/strict';
import test from 'node:test';

import {
  personaCredentials,
  personaSmokeEnv,
  resolveDatabaseURL,
  rolePermissionPlan,
  seedPersonas,
  validateRequiredPasswords,
  validateUniqueUsernames,
} from './seed_bank_soal_role_e2e.js';

test('Bank Soal E2E seed covers the browser smoke personas', () => {
  assert.deepEqual(seedPersonas.map((persona) => persona.key), [
    'admin',
    'creator',
    'reviewer',
    'importer',
    'readonly',
    'no-access',
  ]);

  for (const persona of seedPersonas) {
    const env = personaSmokeEnv(persona);
    assert.equal(env.username, `WEB_ADMIN_BANK_SOAL_E2E_${persona.envPrefix}_USERNAME`);
    assert.equal(env.password, `WEB_ADMIN_BANK_SOAL_E2E_${persona.envPrefix}_PASSWORD`);
  }
});

test('Bank Soal E2E password must be provided via env and is overridable', () => {
  assert.deepEqual(personaCredentials('creator', { BANK_SOAL_E2E_USERNAME_PREFIX: 'seed_' }), {
    username: 'seed_creator',
    password: undefined,
  });

  assert.deepEqual(personaCredentials('creator', {
    BANK_SOAL_E2E_USERNAME_PREFIX: 'seed_',
    BANK_SOAL_E2E_PASSWORD: 'SharedOverride123!',
  }), {
    username: 'seed_creator',
    password: 'SharedOverride123!',
  });

  assert.deepEqual(personaCredentials('creator', {
    BANK_SOAL_E2E_CREATOR_USERNAME: 'custom_creator',
    BANK_SOAL_E2E_CREATOR_PASSWORD: 'CreatorOverride123!',
    BANK_SOAL_E2E_PASSWORD: 'SharedOverride123!',
  }), {
    username: 'custom_creator',
    password: 'CreatorOverride123!',
  });
});

test('Bank Soal E2E role permissions are least-privilege for role smoke', () => {
  const plan = rolePermissionPlan();

  assert.equal(plan.admin.roleCode, 'admin');
  assert.equal(plan.admin.manageRolePermissions, false);

  assert.deepEqual(plan.creator.permissions, [
    'bank_soal.analytics',
    'bank_soal.create',
    'bank_soal.read',
    'bank_soal.update',
  ]);
  assert.ok(!plan.creator.permissions.includes('bank_soal.review'));
  assert.ok(!plan.creator.permissions.includes('bank_soal.import'));
  assert.ok(!plan.creator.permissions.includes('bank_soal.settings'));

  assert.deepEqual(plan.reviewer.permissions, [
    'bank_soal.analytics',
    'bank_soal.read',
    'bank_soal.review',
  ]);
  assert.deepEqual(plan.importer.permissions, [
    'bank_soal.analytics',
    'bank_soal.import',
    'bank_soal.read',
  ]);
  assert.deepEqual(plan.readonly.permissions, [
    'bank_soal.analytics',
    'bank_soal.read',
  ]);
  assert.equal(plan['no-access'].permissions.some((permission) => permission.startsWith('bank_soal.')), false);
});

test('Bank Soal E2E seed requires an explicit database URL by default', () => {
  assert.throws(
    () => resolveDatabaseURL({}),
    /DATABASE_URL is required/,
  );
  assert.equal(
    resolveDatabaseURL({ ALLOW_LOCAL_DATABASE_URL: 'true', LOCAL_DATABASE_URL: 'postgresql://local.example/db' }),
    'postgresql://local.example/db',
  );
  assert.equal(
    resolveDatabaseURL({ DATABASE_URL: 'postgresql://example/db' }),
    'postgresql://example/db',
  );
});

test('Bank Soal E2E seed requires password env before DB writes', () => {
  assert.throws(
    () => validateRequiredPasswords({}),
    /BANK_SOAL_E2E_PASSWORD or per-persona password env is required/,
  );
  assert.doesNotThrow(() => validateRequiredPasswords({ BANK_SOAL_E2E_PASSWORD: 'SharedOverride123!' }));
});

test('Bank Soal E2E seed rejects duplicate persona usernames', () => {
  assert.doesNotThrow(() => validateUniqueUsernames({ BANK_SOAL_E2E_USERNAME_PREFIX: 'seed_' }));
  assert.throws(
    () => validateUniqueUsernames({
      BANK_SOAL_E2E_ADMIN_USERNAME: 'same_user',
      BANK_SOAL_E2E_CREATOR_USERNAME: 'same_user',
    }),
    /duplicate Bank Soal E2E username "same_user" for admin and creator/,
  );
});
