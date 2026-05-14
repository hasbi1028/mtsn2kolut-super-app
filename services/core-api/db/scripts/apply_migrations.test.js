import assert from 'node:assert/strict';
import test from 'node:test';

import {
  assertAppliedChecksum,
  baselineOnExistingSchemaEnabled,
  checksumFor,
  legacyMigrationVersions,
  migrationExecutionPlan,
  resolveDatabaseURL,
  splitSqlStatements,
  stripSqlComments,
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

test('migration execution defaults to transactional and rejects unmarked concurrent indexes', () => {
  assert.equal(
    migrationExecutionPlan('097_safe.sql', 'CREATE INDEX idx_safe ON users(username);').mode,
    'transactional',
  );
  assert.throws(
    () => migrationExecutionPlan('098_missing_marker.sql', 'CREATE INDEX CONCURRENTLY idx_users_username_trgm ON users USING gin (username gin_trgm_ops);'),
    /requires "-- mtsn2kolut:migration non-transactional"/,
  );
});

test('non-transactional migration marker only allows concurrent index statements', () => {
  const sql = `
-- mtsn2kolut:migration non-transactional
DROP INDEX CONCURRENTLY IF EXISTS idx_users_username_trgm;
CREATE INDEX CONCURRENTLY idx_users_username_trgm ON users USING gin (username gin_trgm_ops);
`;
  const plan = migrationExecutionPlan('098_online_index.sql', sql);
  assert.equal(plan.mode, 'nontransactional');
  assert.equal(plan.statements.length, 2);
  assert.match(plan.statements[0], /DROP INDEX CONCURRENTLY/);
  assert.match(plan.statements[1], /CREATE INDEX CONCURRENTLY/);

  assert.throws(
    () => migrationExecutionPlan('098_bad_marker.sql', '-- mtsn2kolut:migration non-transactional\nCREATE TABLE bad(id int);'),
    /reserved for online CREATE\/DROP INDEX CONCURRENTLY/,
  );
  assert.throws(
    () => migrationExecutionPlan('098_bad_statement.sql', '-- mtsn2kolut:migration non-transactional\nBEGIN;\nCREATE INDEX CONCURRENTLY idx_users_username_trgm ON users USING gin (username gin_trgm_ops);'),
    /transaction control is not allowed/,
  );
});

test('SQL statement splitter ignores semicolons inside strings comments and dollar quotes', () => {
  const sql = `
-- ignore ; here
DROP INDEX CONCURRENTLY IF EXISTS idx_one;
CREATE INDEX CONCURRENTLY idx_one ON example USING gin ((regexp_replace(name, ';', '', 'g')) gin_trgm_ops);
/* ignore ; here too */
CREATE INDEX CONCURRENTLY idx_two ON example USING gin ((json_extract_path_text(payload, 'a;b')) gin_trgm_ops);
`;
  const statements = splitSqlStatements(sql);
  assert.equal(statements.length, 3);
  assert.ok(statements.every((statement) => stripSqlComments(statement).includes('CONCURRENTLY')));
});
