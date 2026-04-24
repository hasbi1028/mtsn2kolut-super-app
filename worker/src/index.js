import 'dotenv/config';
import fs from 'fs';
import path from 'path';
import Database from 'better-sqlite3';
import { chromium } from 'playwright';

const DB_PATH = process.env.DB_PATH || path.resolve('../data/pusaka.sqlite');
const POLL_MS = Number(process.env.POLL_MS || 8000);
const MAX_ATTEMPTS_DEFAULT = Math.max(1, Number(process.env.MAX_ATTEMPTS || 3));
const RETRY_BASE_SECONDS = Math.max(10, Number(process.env.RETRY_BASE_SECONDS || 30));
const WORKER_ID_PREFIX = process.env.WORKER_ID || `worker-${process.pid}`;
const LOG_PATH = process.env.WORKER_LOG_PATH || path.resolve('../logs/worker.log');

fs.mkdirSync(path.dirname(LOG_PATH), { recursive: true });

const db = new Database(DB_PATH);
ensureJobColumns();
ensureSettingsTable();

runLoop();

async function runLoop() {
  const initial = loadRuntimeSettings();
  log('INFO', 'worker started', { DB_PATH, POLL_MS, MAX_ATTEMPTS_DEFAULT, ...initial });

  let spawnedCount = 0;

  function spawnConsumer(id) {
    consumerLoop(id).catch((e) => {
      log('ERROR', 'consumer crashed', { id, error: e?.message || String(e) });
    });
  }

  for (let i = 0; i < initial.max_concurrent; i++) {
    spawnConsumer(`${WORKER_ID_PREFIX}-c${++spawnedCount}`);
  }

  // Reload settings every 30s — spawn additional consumers if max_concurrent increased
  while (true) {
    await sleep(30_000);
    try {
      const settings = loadRuntimeSettings();
      if (settings.max_concurrent > spawnedCount) {
        const toAdd = settings.max_concurrent - spawnedCount;
        log('INFO', 'settings changed: spawning consumers', { prev: spawnedCount, next: settings.max_concurrent });
        for (let i = 0; i < toAdd; i++) {
          spawnConsumer(`${WORKER_ID_PREFIX}-c${++spawnedCount}`);
        }
      }
    } catch (e) {
      log('ERROR', 'settings reload failed', { error: e?.message || String(e) });
    }
  }
}

async function consumerLoop(consumerId) {
  while (true) {
    try {
      const claimed = claimNextJob(consumerId);
      if (!claimed) {
        await sleep(POLL_MS);
        continue;
      }

      await processJob(claimed, consumerId);
    } catch (e) {
      log('ERROR', 'consumer loop error', { consumerId, error: e?.message || String(e) });
      await sleep(Math.max(1000, Math.floor(POLL_MS / 2)));
    }
  }
}

function claimNextJob(consumerId) {
  const tx = db.transaction(() => {
    const job = db
      .prepare(
        `SELECT j.*, e.pusaka_username, e.pusaka_password
         FROM jobs j
         JOIN employees e ON e.id = j.employee_id
         WHERE j.status = 'queued'
           AND (j.next_retry_at IS NULL OR datetime(j.next_retry_at) <= CURRENT_TIMESTAMP)
         ORDER BY j.created_at ASC
         LIMIT 1`
      )
      .get();

    if (!job) return null;

    const updated = db
      .prepare(
        `UPDATE jobs
         SET status='running',
             claimed_by=?,
             claimed_at=CURRENT_TIMESTAMP,
             updated_at=CURRENT_TIMESTAMP
         WHERE id=? AND status='queued'`
      )
      .run(consumerId, job.id);

    if (updated.changes !== 1) return null;
    return { ...job, claimed_by: consumerId };
  });

  return tx();
}

async function processJob(job, consumerId) {
  try {
    const record = await scrapeToday(job.pusaka_username, job.pusaka_password);

    db.prepare(
      `INSERT INTO attendance_records (id, employee_id, tanggal, jam_masuk, jam_pulang, source_job_id)
       VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?, ?)
       ON CONFLICT(employee_id, tanggal) DO UPDATE SET
         jam_masuk=excluded.jam_masuk,
         jam_pulang=excluded.jam_pulang,
         source_job_id=excluded.source_job_id,
         updated_at=CURRENT_TIMESTAMP`
    ).run(job.employee_id, record.tanggal, record.jam_masuk, record.jam_pulang, job.id);

    db.prepare(
      `UPDATE jobs
       SET status='success',
           error_message='',
           next_retry_at=NULL,
           updated_at=CURRENT_TIMESTAMP
       WHERE id = ?`
    ).run(job.id);

    log('INFO', 'job success', { job_id: job.id, consumerId, employee_id: job.employee_id, record });
  } catch (e) {
    const errMsg = String(e?.message || e);
    const nextAttempts = Number(job.attempts || 0) + 1;
    const maxAttempts = Number(job.max_attempts || MAX_ATTEMPTS_DEFAULT);

    if (nextAttempts < maxAttempts) {
      const delaySeconds = RETRY_BASE_SECONDS * Math.pow(2, nextAttempts - 1);
      db.prepare(
        `UPDATE jobs
         SET status='queued',
             attempts=?,
             error_message=?,
             next_retry_at=datetime(CURRENT_TIMESTAMP, '+' || ? || ' seconds'),
             updated_at=CURRENT_TIMESTAMP
         WHERE id=?`
      ).run(nextAttempts, errMsg, delaySeconds, job.id);

      log('WARN', 'job queued for retry', {
        job_id: job.id,
        consumerId,
        attempts: nextAttempts,
        maxAttempts,
        retry_in_seconds: delaySeconds,
        error: errMsg
      });
      return;
    }

    db.prepare(
      `UPDATE jobs
       SET status='failed',
           attempts=?,
           error_message=?,
           next_retry_at=NULL,
           updated_at=CURRENT_TIMESTAMP
       WHERE id = ?`
    ).run(nextAttempts, errMsg, job.id);

    log('ERROR', 'job failed permanently', {
      job_id: job.id,
      consumerId,
      attempts: nextAttempts,
      maxAttempts,
      error: errMsg
    });
  }
}

const SCRAPE_RETRIES   = Math.max(1, Number(process.env.SCRAPE_RETRIES   || 3));
const SCRAPE_RETRY_MS  = Math.max(3000, Number(process.env.SCRAPE_RETRY_MS  || 5000));
const ACTION_TIMEOUT   = Math.max(5000, Number(process.env.ACTION_TIMEOUT   || 20000));
const SCREENSHOT_DIR   = process.env.SCREENSHOT_DIR || path.resolve('../logs/screenshots');

async function scrapeToday(username, password) {
  let lastError;
  for (let attempt = 1; attempt <= SCRAPE_RETRIES; attempt++) {
    try {
      return await scrapeOnce(username, password, attempt);
    } catch (e) {
      lastError = e;
      log('WARN', `scrape attempt ${attempt}/${SCRAPE_RETRIES} failed`, { error: e?.message || String(e) });
      if (attempt < SCRAPE_RETRIES) await sleep(SCRAPE_RETRY_MS * attempt);
    }
  }
  throw lastError;
}

async function scrapeOnce(username, password, attempt) {
  const { headless } = loadRuntimeSettings();

  // [3] Browser flags for Linux server stability
  const browser = await chromium.launch({
    headless,
    args: [
      '--no-sandbox',
      '--disable-dev-shm-usage',
      '--disable-gpu',
      '--disable-extensions',
      '--disable-background-networking',
    ],
  });

  // [4] Realistic viewport + User-Agent
  const context = await browser.newContext({
    timezoneId: 'Asia/Makassar',
    viewport: { width: 1280, height: 800 },
    userAgent:
      'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36',
  });

  // [5] Global action timeout — fail-fast per step
  context.setDefaultTimeout(ACTION_TIMEOUT);

  const page = await context.newPage();

  // [2] Block heavy assets — images, fonts, media
  await page.route('**/*', (route) => {
    const type = route.request().resourceType();
    if (['image', 'font', 'media', 'stylesheet'].includes(type)) return route.abort();
    route.continue();
  });

  try {
    // [1] domcontentloaded instead of networkidle
    await page.goto('https://pusaka-v3.kemenag.go.id/', {
      waitUntil: 'domcontentloaded',
      timeout: 45000,
    });

    const loginLink = page.getByRole('link', { name: /login/i }).first();
    await loginLink.waitFor({ state: 'visible' });
    await loginLink.click();

    await page.getByPlaceholder('Username').fill(username);
    await page.getByPlaceholder('Password').fill(password);
    await page.getByRole('button', { name: 'Masuk', exact: true }).click();

    // [8] Explicit login failure detection
    const loginErr = page.getByText(/username atau password salah|invalid credentials|login gagal/i).first();
    const absensiLink = page.getByRole('link', { name: /Absensi/i }).first();
    const which = await Promise.race([
      loginErr.waitFor({ state: 'visible' }).then(() => 'err'),
      absensiLink.waitFor({ state: 'visible' }).then(() => 'ok'),
    ]);
    if (which === 'err') throw new Error('Login gagal: username/password ditolak oleh server');

    await absensiLink.click();

    const riwayatBtn = page.getByRole('button', { name: 'Riwayat Presensi', exact: true });
    await riwayatBtn.waitFor({ state: 'visible' });
    await riwayatBtn.click();

    await page.getByText(/Jam Masuk|Sedang mengambil riwayat/i).first().waitFor().catch(() => {});
    await page.waitForTimeout(2000);

    const bodyText = await page.locator('body').innerText();
    const todayLabel = getTodayLabelID();

    const parsed = parseTodayFromText(bodyText, todayLabel);
    if (!parsed) {
      throw new Error(`Data untuk hari ini tidak ditemukan: ${todayLabel}`);
    }

    return parsed;
  } catch (e) {
    // [7] Screenshot on failure for debugging
    try {
      fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
      const ts   = new Date().toISOString().replace(/[:.]/g, '-');
      const file = path.join(SCREENSHOT_DIR, `fail-a${attempt}-${ts}.png`);
      await page.screenshot({ path: file, fullPage: false });
      log('WARN', 'failure screenshot saved', { file });
    } catch { /* best-effort */ }
    throw e;
  } finally {
    await context.close();
    await browser.close();
  }
}

function parseTodayFromText(text, todayLabel) {
  const normalized = text.replace(/\r/g, '');
  const lines = normalized.split('\n').map((line) => line.trim()).filter(Boolean);
  const candidateIndexes = [];
  for (let i = 0; i < lines.length; i += 1) {
    if (lines[i] === todayLabel) candidateIndexes.push(i);
  }
  if (candidateIndexes.length === 0) return null;

  const idx = candidateIndexes[candidateIndexes.length - 1];

  const blockLines = [];
  for (let i = idx + 1; i < lines.length; i += 1) {
    const line = lines[i];
    if (/^(Senin|Selasa|Rabu|Kamis|Jumat|Sabtu|Minggu),\s+\d{2}\s+\w+\s+\d{4}$/.test(line)) break;
    blockLines.push(line);
  }

  const block = blockLines.join('\n');
  const jamMasuk = extractJam(block, 'Jam Masuk');
  const jamPulang = extractJam(block, 'Jam Pulang');

  return {
    tanggal: toISODateMakassar(),
    jam_masuk: jamMasuk === '-' ? '' : jamMasuk,
    jam_pulang: jamPulang === '-' ? '' : jamPulang
  };
}

function extractJam(block, label) {
  const lines = block.split('\n').map((line) => line.trim()).filter(Boolean);
  for (let i = 0; i < lines.length; i += 1) {
    const line = lines[i];
    if (!line.toLowerCase().includes(label.toLowerCase())) continue;
    const inline = line.match(/([0-9]{2}:[0-9]{2}(?::[0-9]{2})?(?:\s*(?:WITA|WIB|WIT))?|-)/i);
    if (inline) return inline[1].replace(/\s+/g, ' ').trim();
    for (let j = i + 1; j < Math.min(lines.length, i + 4); j += 1) {
      const next = lines[j].match(/([0-9]{2}:[0-9]{2}(?::[0-9]{2})?(?:\s*(?:WITA|WIB|WIT))?|-)/i);
      if (next) return next[1].replace(/\s+/g, ' ').trim();
    }
  }
  return '';
}

function getTodayLabelID() {
  return new Intl.DateTimeFormat('id-ID', {
    weekday: 'long',
    day: '2-digit',
    month: 'long',
    year: 'numeric',
    timeZone: 'Asia/Makassar'
  }).format(new Date());
}

function toISODateMakassar() {
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Makassar',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }).format(new Date());
}

function ensureSettingsTable() {
  db.exec(`
    CREATE TABLE IF NOT EXISTS app_settings (
      key TEXT PRIMARY KEY,
      value TEXT NOT NULL DEFAULT '',
      updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
  `);
}

function loadRuntimeSettings() {
  const rows = db.prepare('SELECT key, value FROM app_settings').all();
  const map = Object.fromEntries(rows.map((row) => [row.key, row.value]));
  return {
    max_concurrent: Math.max(1, Number(map.max_concurrent || process.env.WORKER_CONCURRENCY || 1)),
    headless: String(map.headless ?? process.env.HEADLESS ?? 'false') === '1' || String(map.headless ?? process.env.HEADLESS ?? 'false') === 'true'
  };
}

function ensureJobColumns() {
  const cols = db.prepare('PRAGMA table_info(jobs)').all().map((c) => c.name);
  if (!cols.includes('claimed_by')) db.exec(`ALTER TABLE jobs ADD COLUMN claimed_by TEXT NOT NULL DEFAULT ''`);
  if (!cols.includes('claimed_at')) db.exec(`ALTER TABLE jobs ADD COLUMN claimed_at DATETIME`);
  if (!cols.includes('attempts')) db.exec(`ALTER TABLE jobs ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0`);
  if (!cols.includes('max_attempts')) db.exec(`ALTER TABLE jobs ADD COLUMN max_attempts INTEGER NOT NULL DEFAULT 3`);
  if (!cols.includes('next_retry_at')) db.exec(`ALTER TABLE jobs ADD COLUMN next_retry_at DATETIME`);
}

function log(level, message, context = {}) {
  const line = `${new Date().toISOString()} [${level}] ${message} ${JSON.stringify(context)}`;
  try {
    fs.appendFileSync(LOG_PATH, `${line}\n`);
  } catch {
    // no-op
  }
  if (level === 'ERROR') console.error(line);
  else if (level === 'WARN') console.warn(line);
  else console.log(line);
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
