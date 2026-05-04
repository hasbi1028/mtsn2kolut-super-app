import test, { afterEach, mock } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

afterEach(() => {
  mock.restoreAll();
});

test('logger redacts secrets and hashes usernames before writing or printing', async () => {
  const writes: string[] = [];
  const logs: string[] = [];
  mock.method(fs, 'appendFileSync', (_: fs.PathOrFileDescriptor, data: string | Uint8Array) => {
    writes.push(String(data));
  });
  mock.method(console, 'log', (line: string) => {
    logs.push(line);
  });

  const { log } = await import('./logger.js');
  log('INFO', 'credential check', {
    username: 'operator@example.invalid',
    pusaka_password: 'secret-password',
    nested: {
      token: 'secret-token',
      cookie: 'session=secret-cookie',
      pusaka_username: 'pegawai-1'
    }
  });

  assert.equal(writes.length, 1);
  assert.equal(logs.length, 1);
  const combined = `${writes[0]}\n${logs[0]}`;
  assert.match(combined, /"username_hash":"[a-f0-9]{12}"/);
  assert.match(combined, /"pusaka_password":"\[REDACTED\]"/);
  assert.match(combined, /"token":"\[REDACTED\]"/);
  assert.match(combined, /"cookie":"\[REDACTED\]"/);
  assert.doesNotMatch(combined, /operator@example\.invalid/);
  assert.doesNotMatch(combined, /secret-password/);
  assert.doesNotMatch(combined, /secret-token/);
  assert.doesNotMatch(combined, /secret-cookie/);
  assert.doesNotMatch(combined, /pegawai-1/);
});
