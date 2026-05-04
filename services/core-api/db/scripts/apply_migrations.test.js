import assert from 'node:assert/strict';
import test from 'node:test';

import {
  assertAppliedChecksum,
  baselineOnExistingSchemaEnabled,
  checksumFor,
  legacyMigrationVersions,
  resolveDatabaseURL,
  truthy,
  validateMigrationOrder,
} from './apply_migrations.js';

test('resolveDatabaseURL requires explicit database url by default', () => {
  assert.throws(
    () => resolveDatabaseURL({}),
    /DATABASE_URL is required/,
  );
});

test('resolveDatabaseURL uses local override only when explicitly allowed', () => {
  assert.equal(
    resolveDatabaseURL({ ALLOW_LOCAL_DATABASE_URL: 'true' }),
    'postgresql://pusaka:pusaka_dev@localhost:5432/pusaka',
  );
  assert.equal(
    resolveDatabaseURL({ DATABASE_URL: 'postgresql://example/db', ALLOW_LOCAL_DATABASE_URL: 'true' }),
    'postgresql://example/db',
  );
});

test('truthy accepts only documented local override values', () => {
  assert.equal(truthy('1'), true);
  assert.equal(truthy('true'), true);
  assert.equal(truthy('yes'), true);
  assert.equal(truthy('on'), false);
  assert.equal(truthy(undefined), false);
});

test('baselineOnExistingSchemaEnabled defaults on and accepts explicit disable', () => {
  assert.equal(baselineOnExistingSchemaEnabled({}), true);
  assert.equal(baselineOnExistingSchemaEnabled({ BASELINE_ON_EXISTING_SCHEMA: 'false' }), false);
  assert.equal(baselineOnExistingSchemaEnabled({ BASELINE_ON_EXISTING_SCHEMA: '0' }), false);
  assert.equal(baselineOnExistingSchemaEnabled({ BASELINE_ON_EXISTING_SCHEMA: 'no' }), false);
});

test('validateMigrationOrder rejects duplicate numeric prefixes', () => {
  assert.doesNotThrow(() => validateMigrationOrder([
    '001_initial_schema.sql',
    '027a_library_foundation.sql',
    '060_cbt_exam_token_hardening.sql',
  ]));
  assert.throws(
    () => validateMigrationOrder(['061_first.sql', '061_second.sql']),
    /duplicate migration prefix 061: 061_first.sql, 061_second.sql/,
  );
});

test('checksum helpers are deterministic and reject drift', () => {
  const checksum = checksumFor('select 1;');
  assert.equal(checksum, checksumFor('select 1;'));
  assert.notEqual(checksum, checksumFor('select 2;'));
  assert.doesNotThrow(() => assertAppliedChecksum('001.sql', '001', checksum, checksum));
  assert.doesNotThrow(() => assertAppliedChecksum('001.sql', '001', '', checksum));
  assert.throws(
    () => assertAppliedChecksum('001.sql', '001', checksum, checksumFor('select 2;')),
    /checksum mismatch for 001.sql \(applied as 001\)/,
  );
});

test('legacy migration versions preserve library rename compatibility', () => {
  assert.deepEqual(legacyMigrationVersions('027a_library_foundation'), ['027_library_foundation']);
  assert.deepEqual(legacyMigrationVersions('060_cbt_exam_token_hardening'), []);
});
