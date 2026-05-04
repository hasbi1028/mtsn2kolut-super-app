import test from 'node:test';
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';

import { normalizeRuntimeConfigPatch } from './config.js';

function importConfig(env: NodeJS.ProcessEnv) {
  return spawnSync(
    process.execPath,
    [
      '--import',
      'tsx',
      '--input-type=module',
      '-e',
      "import('./src/config.ts').then(() => process.exit(0)).catch((error) => { console.error(error.message); process.exit(42); })",
    ],
    {
      cwd: new URL('..', import.meta.url),
      env: {
        ...process.env,
        ...env,
      },
      encoding: 'utf8',
    },
  );
}

test('production-like config requires WORKER_API_KEY without leaking values', () => {
  const result = importConfig({
    NODE_ENV: 'production',
    WORKER_ENV: 'production',
    APP_ENV: 'production',
    BACKEND_URL: 'https://api.example.invalid/sensitive/path',
    WORKER_API_KEY: '',
  });

  assert.equal(result.status, 42);
  assert.match(result.stderr, /WORKER_API_KEY is required/);
  assert.doesNotMatch(result.stderr, /api\.example\.invalid/);
});

test('numeric config rejects invalid values with sanitized message', () => {
  const result = importConfig({
    NODE_ENV: 'test',
    BACKEND_URL: 'http://localhost:8080',
    WORKER_CONCURRENCY: 'NaN',
  });

  assert.equal(result.status, 42);
  assert.match(result.stderr, /WORKER_CONCURRENCY must be a finite number/);
  assert.doesNotMatch(result.stderr, /NaN/);
});

test('normalizeRuntimeConfigPatch bounds backend maxConcurrent', () => {
  assert.deepEqual(
    normalizeRuntimeConfigPatch({ max_concurrent: '50', headless: 'yes' }),
    { maxConcurrent: 50, headless: true },
  );
  assert.deepEqual(
    normalizeRuntimeConfigPatch({ max_concurrent: '51', headless: 'false' }),
    { headless: false },
  );
  assert.deepEqual(normalizeRuntimeConfigPatch({ max_concurrent: '0' }), {});
  assert.deepEqual(normalizeRuntimeConfigPatch({ maxConcurrent: 3.9 }), {
    maxConcurrent: 3,
  });
});
