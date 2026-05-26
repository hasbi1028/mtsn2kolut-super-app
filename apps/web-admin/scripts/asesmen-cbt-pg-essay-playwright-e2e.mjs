#!/usr/bin/env node
import { spawn, spawnSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium, request } from 'playwright';

const __dirname = dirname(fileURLToPath(import.meta.url));
const appDir = resolve(__dirname, '..');
const repoRoot = resolve(appDir, '../..');
const coreDir = resolve(repoRoot, 'services/core-api');

const apiPort = Number(process.env.E2E_API_PORT ?? 18080);
const webPort = Number(process.env.E2E_WEB_PORT ?? 18081);
const apiBase = process.env.E2E_API_BASE ?? `http://127.0.0.1:${apiPort}`;
const webBase = process.env.E2E_WEB_BASE ?? `http://127.0.0.1:${webPort}`;
const adminPassword = process.env.E2E_ADMIN_PASSWORD ?? 'E2eAdmin#2026!';
const jwtSecret = process.env.JWT_SECRET ?? 'e2e-test-jwt-secret';
const studentToken = process.env.E2E_CBT_TOKEN ?? 'a1b2c3d4';
const roomToken = process.env.E2E_ROOM_TOKEN ?? 'ROOM2501';
const deviceFingerprint = `pw-asesmen-pg-essay-${Date.now()}`;

const children = [];
let browser;
let apiContext;

function log(message) {
  console.log(`[asesmen-cbt-pg-essay-e2e] ${message}`);
}

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd ?? repoRoot,
    env: { ...process.env, ...(options.env ?? {}) },
    encoding: 'utf8',
    stdio: options.stdio ?? 'pipe',
  });
  if (result.status !== 0) {
    throw new Error(`${command} ${args.join(' ')} failed\n${result.stdout ?? ''}\n${result.stderr ?? ''}`);
  }
  return result.stdout ?? '';
}

function testDatabaseUrl() {
  const out = run('bash', ['-lc', `set -a; . services/core-api/.env; set +a; python3 - <<'PY'
import os
from urllib.parse import urlsplit, urlunsplit
u=os.environ['DATABASE_URL']
p=urlsplit(u)
db=p.path.rsplit('/', 1)[-1]
if db.endswith('_test'):
    test_db = db
else:
    test_db = db + '_test'
print(urlunsplit(p._replace(path='/' + test_db)))
PY`]);
  const value = out.trim();
  if (!value || !value.includes('_test')) throw new Error(`Refusing to run without *_test database URL: ${value}`);
  return value;
}

async function waitFor(url, label, timeoutMs = 60_000) {
  const deadline = Date.now() + timeoutMs;
  let lastError;
  while (Date.now() < deadline) {
    try {
      const res = await fetch(url, { headers: { accept: 'application/json' } });
      if (res.ok) return;
      lastError = new Error(`${label} returned ${res.status}`);
    } catch (error) {
      lastError = error;
    }
    await new Promise((resolveWait) => setTimeout(resolveWait, 750));
  }
  throw new Error(`${label} not ready at ${url}: ${lastError?.message ?? 'timeout'}`);
}

function startProcess(command, args, options) {
  const child = spawn(command, args, {
    cwd: options.cwd,
    env: { ...process.env, ...options.env },
    stdio: ['ignore', 'pipe', 'pipe'],
    detached: true,
  });
  children.push(child);
  child.stdout.on('data', (chunk) => process.stdout.write(`[${options.name}] ${chunk}`));
  child.stderr.on('data', (chunk) => process.stderr.write(`[${options.name}] ${chunk}`));
  child.on('exit', (code) => {
    if (code !== null && code !== 0 && !options.allowExit) {
      console.error(`[${options.name}] exited with code ${code}`);
    }
  });
  return child;
}

async function apiJson(method, path, body, headers = {}) {
  const response = await apiContext.fetch(`${apiBase}${path}`, {
    method,
    headers: { accept: 'application/json', ...(body ? { 'content-type': 'application/json' } : {}), ...headers },
    data: body,
  });
  const text = await response.text();
  let json;
  try {
    json = JSON.parse(text);
  } catch {
    json = undefined;
  }
  if (!response.ok()) {
    throw new Error(`${method} ${path} failed with ${response.status()} ${text}`);
  }
  return json;
}

async function cleanup() {
  if (apiContext) await apiContext.dispose().catch(() => {});
  if (browser) await browser.close().catch(() => {});
  for (const child of children.reverse()) {
    try {
      process.kill(-child.pid, 'SIGTERM');
    } catch {
      if (!child.killed) child.kill('SIGTERM');
    }
  }
}

async function main() {
  const databaseUrl = testDatabaseUrl();
  log(`Using isolated test DB: ${databaseUrl.replace(/:([^:@/]+)@/, ':***@')}`);

  if (!existsSync(resolve(coreDir, 'bin/api'))) {
    log('Building core-api binary...');
    run('go', ['build', '-o', 'bin/api', './cmd/api'], { cwd: coreDir, stdio: 'inherit' });
  }

  log('Resetting deterministic CBT seed...');
  run('bash', ['scripts/setup-cbt-test-db.sh', '--no-migrate'], { cwd: repoRoot, stdio: 'inherit' });
  run('psql', [databaseUrl, '-v', 'ON_ERROR_STOP=1', '-c', "UPDATE users SET must_change_password = FALSE, password_changed_at = NOW() WHERE username = 'admin';"], { cwd: repoRoot });

  log(`Starting core-api on ${apiBase}...`);
  startProcess('bash', ['-lc', 'set -a; . services/core-api/.env; set +a; DATABASE_URL="$E2E_DATABASE_URL" PORT="$E2E_API_PORT" ADMIN_PASSWORD="$E2E_ADMIN_PASSWORD" JWT_SECRET="$E2E_JWT_SECRET" ./services/core-api/bin/api'], {
    name: 'core-api-e2e',
    cwd: repoRoot,
    env: { E2E_DATABASE_URL: databaseUrl, E2E_API_PORT: String(apiPort), E2E_ADMIN_PASSWORD: adminPassword, E2E_JWT_SECRET: jwtSecret },
  });
  await waitFor(`${apiBase}/health`, 'core-api');

  log(`Starting web-admin dev server on ${webBase}...`);
  startProcess('npm', ['run', 'dev', '--', '--host', '127.0.0.1', '--port', String(webPort)], {
    name: 'web-admin-e2e',
    cwd: appDir,
    env: { API_BASE_URL: apiBase, PORT: String(webPort), NODE_ENV: 'test' },
  });
  await waitFor(`${webBase}/ujian`, 'web-admin /ujian');

  apiContext = await request.newContext({ baseURL: apiBase });

  log('Running browser flow: /ujian legacy token login, PG + essay answers, submit...');
  browser = await chromium.launch({ headless: true, args: ['--no-sandbox'] });
  const page = await browser.newPage();
  page.on('dialog', (dialog) => dialog.accept());
  await page.goto(`${webBase}/ujian`, { waitUntil: 'networkidle' });
  await page.getByRole('button', { name: /mode bantuan pengawas/i }).click();
  await page.getByLabel(/Token Ujian/i).fill(studentToken);
  await page.getByLabel(/Token Ruang/i).fill(roomToken);
  await page.getByRole('button', { name: /Lanjutkan/i }).click();
  await page.getByRole('heading', { name: /Apakah data ini benar/i }).waitFor({ timeout: 30_000 });
  await page.getByRole('button', { name: /Ya, Masuk Ujian/i }).click();
  await page.getByText(/Soal 1/i).waitFor({ timeout: 30_000 });

  await page.locator('article').filter({ hasText: '12 x 8' }).locator('label', { hasText: /^B\.\s*96$|96/ }).locator('input').check();
  await page.locator('article').filter({ hasText: 'total buku' }).locator('label', { hasText: /96 buku/ }).locator('input').check();
  await page.locator('article').filter({ hasText: 'Jaringan internet yang stabil' }).locator('label', { hasText: /Benar/ }).locator('input').check();
  const essayArticle = page.locator('article').filter({ hasText: 'koneksi aplikasi ujian menurun' });
  await essayArticle.getByPlaceholder(/Tulis jawaban/i).fill('Tetap di aplikasi, lapor ke pengawas, dan menunggu status sinkron pulih.');
  await essayArticle.getByPlaceholder(/Tulis jawaban/i).blur();
  await page.getByText(/4\/4 terjawab/).waitFor({ timeout: 30_000 });
  await page.getByRole('button', { name: /Kumpulkan/i }).click();
  await page.getByRole('heading', { name: /^Selesai$/i }).waitFor({ timeout: 30_000 });

  log('Running nilai flow through API: score PG, grade essay, verify result/preflight...');
  const adminLogin = await apiJson('POST', '/api/auth/login', { username: 'admin', password: adminPassword });
  const adminHeaders = { authorization: `Bearer ${adminLogin.data.access_token}` };

  const portalLogin = await apiJson('POST', '/api/cbt-portal/login', { nisn: '9990000001', code: '9990000001' });
  const participant = portalLogin.data.schedule.find((item) => item.can_start || item.status === 'submitted') ?? portalLogin.data.schedule[0];
  if (!participant?.session_id || !participant?.participant_id) throw new Error('Seed participant/session not found through portal schedule');

  await apiJson('POST', `/api/asesmen/sessions/${participant.session_id}/score`, {}, adminHeaders);
  const ungraded = await apiJson('GET', `/api/asesmen/sessions/${participant.session_id}/ungraded-essays`, undefined, adminHeaders);
  const essay = ungraded.data.find((item) => item.participant_id === participant.participant_id);
  if (!essay?.answer_id) throw new Error(`Expected one ungraded essay for participant ${participant.participant_id}`);
  await apiJson('POST', `/api/asesmen/sessions/${participant.session_id}/answers/${essay.answer_id}/grade-essay`, { manual_score: 80 }, adminHeaders);

  const results = await apiJson('GET', `/api/asesmen/sessions/${participant.session_id}/results`, undefined, adminHeaders);
  const row = results.data.results.find((item) => item.participant_id === participant.participant_id);
  if (!row) throw new Error('Participant result row missing');
  if (row.total_answers !== 4) throw new Error(`Expected 4 answers, got ${row.total_answers}`);
  if (row.correct_answers !== 3) throw new Error(`Expected 3 correct PG answers, got ${row.correct_answers}`);
  if (Number(row.score) !== 95) throw new Error(`Expected final score 95, got ${row.score}`);

  const preflight = await apiJson('GET', `/api/asesmen/sessions/${participant.session_id}/grade-sync-preflight`, undefined, adminHeaders);
  if (preflight.data.missing_score_count !== 0) throw new Error(`Expected missing_score_count 0, got ${preflight.data.missing_score_count}`);
  if (!preflight.data.grade_assignment_id) throw new Error('Expected grade_assignment_id to be present for Nilai sync readiness');

  log(JSON.stringify({ status: 'passed', participant_id: participant.participant_id, score: row.score, total_answers: row.total_answers, correct_answers: row.correct_answers, grade_assignment_id: preflight.data.grade_assignment_id }, null, 2));
}

main()
  .catch((error) => {
    console.error(error);
    process.exitCode = 1;
  })
  .finally(cleanup);
