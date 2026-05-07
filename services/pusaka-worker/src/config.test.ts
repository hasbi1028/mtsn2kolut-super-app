import test from 'node:test';
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';

import { normalizeRuntimeConfigPatch } from './config.js';

function importConfig(env: NodeJS.ProcessEnv, script = "import('./src/config.ts').then(() => process.exit(0)).catch((error) => { console.error(error.message); process.exit(42); })") {
  return spawnSync(
    process.execPath,
    [
      '--import',
      'tsx',
      '--input-type=module',
      '-e',
      script,
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

test('production-like config requires explicit BACKEND_URL without falling back to localhost', () => {
  const result = importConfig({
    NODE_ENV: 'production',
    WORKER_ENV: 'production',
    APP_ENV: 'production',
    BACKEND_URL: '',
    WORKER_API_KEY: 'worker-key',
  });

  assert.equal(result.status, 42);
  assert.match(result.stderr, /BACKEND_URL is required/);
  assert.doesNotMatch(result.stderr, /worker-key/);
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

test('BACKEND_URL rejects credentials, query, and fragment', () => {
  for (const BACKEND_URL of [
    'https://worker:secret@api.example.invalid',
    'https://api.example.invalid?token=secret',
    'https://api.example.invalid#secret',
  ]) {
    const result = importConfig({ NODE_ENV: 'test', BACKEND_URL });

    assert.equal(result.status, 42);
    assert.match(result.stderr, /BACKEND_URL must not include/);
    assert.doesNotMatch(result.stderr, /secret/);
  }
});

test('BACKEND_URL requires http or https and normalizes trailing slash safely', () => {
  const invalid = importConfig({ NODE_ENV: 'test', BACKEND_URL: 'ftp://api.example.invalid' });
  assert.equal(invalid.status, 42);
  assert.match(invalid.stderr, /BACKEND_URL must use http or https/);

  const valid = importConfig(
    { NODE_ENV: 'test', BACKEND_URL: 'https://api.example.invalid/pusaka/' },
    "import('./src/config.ts').then((config) => { console.log(config.BACKEND_URL); process.exit(0); }).catch((error) => { console.error(error.message); process.exit(42); })",
  );
  assert.equal(valid.status, 0);
  assert.equal(valid.stdout.trim(), 'https://api.example.invalid/pusaka');
});

test('worker heartbeat cadence is configured independently from config sync', () => {
  const result = importConfig(
    {
      NODE_ENV: 'test',
      BACKEND_URL: 'http://localhost:8080',
      CONFIG_SYNC_MS: '300000',
      WORKER_HEARTBEAT_MS: '30000',
    },
    "import('./src/config.ts').then((config) => { console.log(`${config.CONFIG_SYNC_MS}:${config.WORKER_HEARTBEAT_MS}`); process.exit(0); }).catch((error) => { console.error(error.message); process.exit(42); })",
  );

  assert.equal(result.status, 0);
  assert.equal(result.stdout.trim(), '300000:30000');
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

test('normalizeRuntimeConfigPatch normalizes geolocation settings', () => {
  assert.deepEqual(
    normalizeRuntimeConfigPatch({
      pusaka_geo_base_lat: '-3.2',
      pusaka_geo_base_lng: '121.1',
      pusaka_geo_default_radius_m: '40',
      pusaka_geo_checkin_radius_m: '60',
      pusaka_geo_checkout_radius_m: '30',
    }),
    {
      geo: {
        baseLat: -3.2,
        baseLng: 121.1,
        defaultRadiusMeters: 40,
        checkinRadiusMeters: 60,
        checkoutRadiusMeters: 30,
      },
    },
  );

  assert.deepEqual(
    normalizeRuntimeConfigPatch({
      pusaka_geo_base_lat: '-91',
      pusaka_geo_base_lng: '181',
      pusaka_geo_checkin_radius_m: '201',
    }),
    {},
  );
});
