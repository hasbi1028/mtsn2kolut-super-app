import 'dotenv/config';
import fs from 'fs';
import path from 'path';
import Database from 'better-sqlite3';
import { chromium } from 'playwright';

const DB_PATH             = process.env.DB_PATH             || path.resolve('../data/pusaka.sqlite');
const POLL_MS             = Number(process.env.POLL_MS             || 8000);
const MAX_ATTEMPTS_DEFAULT= Math.max(1, Number(process.env.MAX_ATTEMPTS       || 3));
const RETRY_BASE_SECONDS  = Math.max(10, Number(process.env.RETRY_BASE_SECONDS || 30));
const WORKER_ID_PREFIX    = process.env.WORKER_ID            || `worker-${process.pid}`;
const LOG_PATH            = process.env.WORKER_LOG_PATH      || path.resolve('../logs/worker.log');
const SCRAPE_RETRIES      = Math.max(1, Number(process.env.SCRAPE_RETRIES     || 3));
const SCRAPE_RETRY_MS     = Math.max(3000, Number(process.env.SCRAPE_RETRY_MS  || 5000));
const ACTION_TIMEOUT      = Math.max(5000, Number(process.env.ACTION_TIMEOUT   || 20000));
const SCREENSHOT_DIR      = process.env.SCREENSHOT_DIR       || path.resolve('../logs/screenshots');

const BASE_URL = 'https://pusaka-v3.kemenag.go.id';
const BASE_LAT = -3.2163111;
const BASE_LNG = 121.0428659;

const BROWSER_ARGS = [
  '--no-sandbox',
  '--disable-dev-shm-usage',
  '--disable-gpu',
  '--disable-extensions',
  '--disable-background-networking',
];

const USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36';

fs.mkdirSync(path.dirname(LOG_PATH), { recursive: true });

const db = new Database(DB_PATH);
ensureJobColumns();
ensureSettingsTable();

runLoop();

// ── Main loop ────────────────────────────────────────────────────────────────

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
    const job = db.prepare(
      `SELECT j.*, e.pusaka_username, e.pusaka_password
       FROM jobs j
       JOIN employees e ON e.id = j.employee_id
       WHERE j.status = 'queued'
         AND (j.next_retry_at IS NULL OR datetime(j.next_retry_at) <= CURRENT_TIMESTAMP)
       ORDER BY j.created_at ASC
       LIMIT 1`
    ).get();

    if (!job) return null;

    const updated = db.prepare(
      `UPDATE jobs
       SET status='running', claimed_by=?, claimed_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP
       WHERE id=? AND status='queued'`
    ).run(consumerId, job.id);

    if (updated.changes !== 1) return null;
    return { ...job, claimed_by: consumerId };
  });

  return tx();
}

async function processJob(job, consumerId) {
  try {
    let record;

    if (job.run_type === 'checkin') {
      await withRetry((attempt) => checkin(job.pusaka_username, job.pusaka_password, attempt), 'checkin');
      record = {
        tanggal:    toISODateMakassar(),
        jam_masuk:  toTimeWITA(),
        jam_pulang: '',
      };
    } else if (job.run_type === 'checkout') {
      await withRetry((attempt) => checkout(job.pusaka_username, job.pusaka_password, attempt), 'checkout');
      record = {
        tanggal:    toISODateMakassar(),
        jam_masuk:  '',
        jam_pulang: toTimeWITA(),
      };
    } else {
      record = await withRetry((attempt) => scrapeOnce(job.pusaka_username, job.pusaka_password, attempt), 'scrape');
    }

    // UPSERT: only overwrite a field if the new value is non-empty,
    // so a checkin job never wipes an existing jam_pulang and vice-versa.
    db.prepare(
      `INSERT INTO attendance_records (id, employee_id, tanggal, jam_masuk, jam_pulang, source_job_id)
       VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?, ?)
       ON CONFLICT(employee_id, tanggal) DO UPDATE SET
         jam_masuk     = CASE WHEN excluded.jam_masuk  != '' THEN excluded.jam_masuk  ELSE jam_masuk  END,
         jam_pulang    = CASE WHEN excluded.jam_pulang != '' THEN excluded.jam_pulang ELSE jam_pulang END,
         source_job_id = excluded.source_job_id,
         updated_at    = CURRENT_TIMESTAMP`
    ).run(job.employee_id, record.tanggal, record.jam_masuk, record.jam_pulang, job.id);

    db.prepare(
      `UPDATE jobs SET status='success', error_message='', next_retry_at=NULL, updated_at=CURRENT_TIMESTAMP
       WHERE id = ?`
    ).run(job.id);

    log('INFO', 'job success', { job_id: job.id, consumerId, employee_id: job.employee_id, record });
  } catch (e) {
    const errMsg      = String(e?.message || e);
    const nextAttempts= Number(job.attempts || 0) + 1;
    const maxAttempts = Number(job.max_attempts || MAX_ATTEMPTS_DEFAULT);

    if (nextAttempts < maxAttempts) {
      const delaySeconds = RETRY_BASE_SECONDS * Math.pow(2, nextAttempts - 1);
      db.prepare(
        `UPDATE jobs
         SET status='queued', attempts=?, error_message=?,
             next_retry_at=datetime(CURRENT_TIMESTAMP, '+' || ? || ' seconds'), updated_at=CURRENT_TIMESTAMP
         WHERE id=?`
      ).run(nextAttempts, errMsg, delaySeconds, job.id);

      log('WARN', 'job queued for retry', { job_id: job.id, consumerId, attempts: nextAttempts, maxAttempts, retry_in_seconds: delaySeconds, error: errMsg });
      return;
    }

    db.prepare(
      `UPDATE jobs
       SET status='failed', attempts=?, error_message=?, next_retry_at=NULL, updated_at=CURRENT_TIMESTAMP
       WHERE id = ?`
    ).run(nextAttempts, errMsg, job.id);

    log('ERROR', 'job failed permanently', { job_id: job.id, consumerId, attempts: nextAttempts, maxAttempts, error: errMsg });
  }
}

// ── Retry wrapper ─────────────────────────────────────────────────────────────

async function withRetry(fn, label) {
  let lastError;
  for (let attempt = 1; attempt <= SCRAPE_RETRIES; attempt++) {
    try {
      return await fn(attempt);
    } catch (e) {
      lastError = e;
      log('WARN', `${label} attempt ${attempt}/${SCRAPE_RETRIES} failed`, { error: e?.message || String(e) });
      if (attempt < SCRAPE_RETRIES) await sleep(SCRAPE_RETRY_MS * attempt);
    }
  }
  throw lastError;
}

// ── Shared browser helpers ────────────────────────────────────────────────────

function getBrowserLaunchOptions(headless) {
  return { headless, args: BROWSER_ARGS };
}

function getBaseContext(extra = {}) {
  return {
    timezoneId: 'Asia/Makassar',
    locale: 'id-ID',
    viewport: { width: 1280, height: 800 },
    userAgent: USER_AGENT,
    ...extra,
  };
}

async function blockAssets(page) {
  await page.route('**/*', (route) => {
    if (['image', 'font', 'media', 'stylesheet'].includes(route.request().resourceType()))
      return route.abort();
    route.continue();
  });
}

async function loginToPusaka(page, username, password, label) {
  await page.goto(BASE_URL, { waitUntil: 'domcontentloaded', timeout: 45000 });

  const loginLink = page.getByRole('link', { name: /login/i }).first();
  await loginLink.waitFor({ state: 'visible' });
  await loginLink.click();

  await page.getByPlaceholder('Username').fill(username);
  await page.getByPlaceholder('Password').fill(password);
  await page.getByRole('button', { name: 'Masuk', exact: true }).click();

  const loginErr   = page.getByText(/username atau password salah|invalid credentials|login gagal/i).first();
  const profileLink= page.getByRole('link', { name: /profile/i }).first();
  const which = await Promise.race([
    loginErr.waitFor({ state: 'visible' }).then(() => 'err'),
    profileLink.waitFor({ state: 'visible', timeout: 20000 }).then(() => 'ok'),
  ]);
  if (which === 'err') throw new Error('Login gagal: username/password ditolak oleh server');
  log('INFO', `${label}: login ok`, { username });
}

async function triggerGeo(page) {
  await page.evaluate(() => new Promise((resolve) => {
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve(pos),
      (err) => resolve(err.message),
      { enableHighAccuracy: true, timeout: 5000 }
    );
  })).catch(() => {});
}

async function assertPresenceResult(page, label, username) {
  const text = await page.locator('body').innerText().catch(() => '');
  if (text.includes('PRESENSI GAGAL'))
    throw new Error('PRESENSI GAGAL — Bad Request (mungkin GPS tidak valid)');
  if (/berhasil/i.test(text))
    log('INFO', `${label}: BERHASIL`, { username });
  else if (/sudah presensi/i.test(text))
    log('WARN', `${label}: sudah absen sebelumnya`, { username });
  else
    log('WARN', `${label}: status tidak jelas, cek screenshot`, { username, snippet: text.substring(0, 300) });
}

async function saveFailScreenshot(page, name) {
  try {
    fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
    const ts   = new Date().toISOString().replace(/[:.]/g, '-');
    const file = path.join(SCREENSHOT_DIR, `${name}-${ts}.png`);
    await page.screenshot({ path: file, fullPage: false });
    log('WARN', 'failure screenshot saved', { file });
  } catch { /* best-effort */ }
}

// ── Scrape (morning / afternoon) ──────────────────────────────────────────────

async function scrapeOnce(username, password, attempt) {
  const { headless } = loadRuntimeSettings();
  const browser = await chromium.launch(getBrowserLaunchOptions(headless));
  const context = await browser.newContext(getBaseContext());
  context.setDefaultTimeout(ACTION_TIMEOUT);
  const page = await context.newPage();
  await blockAssets(page);

  try {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded', timeout: 45000 });

    const loginLink = page.getByRole('link', { name: /login/i }).first();
    await loginLink.waitFor({ state: 'visible' });
    await loginLink.click();

    await page.getByPlaceholder('Username').fill(username);
    await page.getByPlaceholder('Password').fill(password);
    await page.getByRole('button', { name: 'Masuk', exact: true }).click();

    const loginErr   = page.getByText(/username atau password salah|invalid credentials|login gagal/i).first();
    const absensiLink= page.getByRole('link', { name: /Absensi/i }).first();
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

    const bodyText   = await page.locator('body').innerText();
    const todayLabel = getTodayLabelID();
    const parsed     = parseTodayFromText(bodyText, todayLabel);
    if (!parsed) throw new Error(`Data untuk hari ini tidak ditemukan: ${todayLabel}`);

    return parsed;
  } catch (e) {
    await saveFailScreenshot(page, `scrape-fail-a${attempt}`);
    throw e;
  } finally {
    await context.close();
    await browser.close();
  }
}

// ── Checkin ───────────────────────────────────────────────────────────────────

async function checkin(username, password, attempt) {
  const label = 'checkin';
  const geo   = randomGeo(BASE_LAT, BASE_LNG, 28);
  const { headless } = loadRuntimeSettings();

  const browser = await chromium.launch(getBrowserLaunchOptions(headless));
  const context = await browser.newContext(getBaseContext({ geolocation: geo, permissions: ['geolocation'] }));
  context.setDefaultTimeout(ACTION_TIMEOUT);
  const page = await context.newPage();

  try {
    log('INFO', `${label}: starting`, { username, lat: geo.latitude.toFixed(7), lng: geo.longitude.toFixed(7) });
    await loginToPusaka(page, username, password, label);

    await page.goto(`${BASE_URL}/profile/presence`, { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(3000);
    await triggerGeo(page);
    await page.waitForTimeout(2000);

    const btn = page.locator('button:has-text("Presensi masuk")');
    if (await btn.count() === 0) {
      const bodyText = await page.locator('body').innerText().catch(() => '');
      if (/hadir|sudah/i.test(bodyText)) {
        log('WARN', `${label}: sudah absen hari ini`, { username });
        return;
      }
      throw new Error('Tombol Presensi masuk tidak ditemukan');
    }
    if (await btn.first().isDisabled().catch(() => false)) {
      log('WARN', `${label}: tombol disabled (sudah absen)`, { username });
      return;
    }

    await triggerGeo(page);
    await page.waitForTimeout(1000);
    await btn.first().click();
    log('INFO', `${label}: tombol diklik`, { username });

    await page.waitForTimeout(3000);
    await assertPresenceResult(page, label, username);
  } catch (e) {
    await saveFailScreenshot(page, `${label}-fail-a${attempt}-${username}`);
    throw e;
  } finally {
    await context.close();
    await browser.close();
  }
}

// ── Checkout ──────────────────────────────────────────────────────────────────

async function checkout(username, password, attempt) {
  const label = 'checkout';
  const geo   = randomGeo(BASE_LAT, BASE_LNG, 28);
  const { headless } = loadRuntimeSettings();

  const browser = await chromium.launch(getBrowserLaunchOptions(headless));
  const context = await browser.newContext(getBaseContext({ geolocation: geo, permissions: ['geolocation'] }));
  context.setDefaultTimeout(ACTION_TIMEOUT);
  const page = await context.newPage();

  try {
    log('INFO', `${label}: starting`, { username, lat: geo.latitude.toFixed(7), lng: geo.longitude.toFixed(7) });
    await loginToPusaka(page, username, password, label);

    await page.goto(`${BASE_URL}/profile/presence`, { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(3000);
    await triggerGeo(page);
    await page.waitForTimeout(2000);

    const btn = page.locator('button:has-text("Presensi pulang")');
    if (await btn.count() === 0) throw new Error('Tombol Presensi pulang tidak ditemukan');
    if (await btn.first().isDisabled().catch(() => false)) {
      log('WARN', `${label}: tombol disabled (sudah absen pulang)`, { username });
      return;
    }

    await btn.first().click();
    log('INFO', `${label}: tombol diklik`, { username });
    await page.waitForTimeout(2000);

    // Konfirmasi modal "Ya"
    const yaBtn = page.getByRole('button', { name: /^ya$/i });
    if (await yaBtn.count() > 0 && !await yaBtn.first().isDisabled().catch(() => false)) {
      await yaBtn.first().click();
      log('INFO', `${label}: konfirmasi Ya diklik`, { username });
    }

    await page.waitForTimeout(3000);
    await assertPresenceResult(page, label, username);
  } catch (e) {
    await saveFailScreenshot(page, `${label}-fail-a${attempt}-${username}`);
    throw e;
  } finally {
    await context.close();
    await browser.close();
  }
}

// ── Geolocation helpers ───────────────────────────────────────────────────────

function randomGeo(baseLat, baseLng, radiusMeters = 10) {
  const R = 6371000;
  const latOff = (Math.random() - 0.5) * 2 * (radiusMeters / R) * (180 / Math.PI);
  const lngOff = ((Math.random() - 0.5) * 2 * (radiusMeters / R) * (180 / Math.PI)) / Math.cos((baseLat * Math.PI) / 180);
  return { latitude: baseLat + latOff, longitude: baseLng + lngOff };
}

// ── Text parsing ──────────────────────────────────────────────────────────────

function parseTodayFromText(text, todayLabel) {
  const lines = text.replace(/\r/g, '').split('\n').map((l) => l.trim()).filter(Boolean);
  const candidates = [];
  for (let i = 0; i < lines.length; i++) {
    if (lines[i] === todayLabel) candidates.push(i);
  }
  if (candidates.length === 0) return null;

  const idx = candidates[candidates.length - 1];
  const blockLines = [];
  for (let i = idx + 1; i < lines.length; i++) {
    if (/^(Senin|Selasa|Rabu|Kamis|Jumat|Sabtu|Minggu),\s+\d{2}\s+\w+\s+\d{4}$/.test(lines[i])) break;
    blockLines.push(lines[i]);
  }

  const block     = blockLines.join('\n');
  const jamMasuk  = extractJam(block, 'Jam Masuk');
  const jamPulang = extractJam(block, 'Jam Pulang');

  return {
    tanggal:    toISODateMakassar(),
    jam_masuk:  jamMasuk  === '-' ? '' : jamMasuk,
    jam_pulang: jamPulang === '-' ? '' : jamPulang,
  };
}

function extractJam(block, label) {
  const lines = block.split('\n').map((l) => l.trim()).filter(Boolean);
  for (let i = 0; i < lines.length; i++) {
    if (!lines[i].toLowerCase().includes(label.toLowerCase())) continue;
    const inline = lines[i].match(/([0-9]{2}:[0-9]{2}(?::[0-9]{2})?(?:\s*(?:WITA|WIB|WIT))?|-)/i);
    if (inline) return inline[1].replace(/\s+/g, ' ').trim();
    for (let j = i + 1; j < Math.min(lines.length, i + 4); j++) {
      const next = lines[j].match(/([0-9]{2}:[0-9]{2}(?::[0-9]{2})?(?:\s*(?:WITA|WIB|WIT))?|-)/i);
      if (next) return next[1].replace(/\s+/g, ' ').trim();
    }
  }
  return '';
}

// ── Date / time helpers ───────────────────────────────────────────────────────

function getTodayLabelID() {
  return new Intl.DateTimeFormat('id-ID', {
    weekday: 'long', day: '2-digit', month: 'long', year: 'numeric',
    timeZone: 'Asia/Makassar',
  }).format(new Date());
}

function toISODateMakassar() {
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Makassar', year: 'numeric', month: '2-digit', day: '2-digit',
  }).format(new Date());
}

function toTimeWITA() {
  return new Intl.DateTimeFormat('id-ID', {
    timeZone: 'Asia/Makassar', hour: '2-digit', minute: '2-digit', hour12: false,
  }).format(new Date());
}

// ── DB helpers ────────────────────────────────────────────────────────────────

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
  const map  = Object.fromEntries(rows.map((r) => [r.key, r.value]));
  return {
    max_concurrent: Math.max(1, Number(map.max_concurrent || process.env.WORKER_CONCURRENCY || 1)),
    headless: ['1', 'true'].includes(String(map.headless ?? process.env.HEADLESS ?? 'false')),
  };
}

function ensureJobColumns() {
  const cols = db.prepare('PRAGMA table_info(jobs)').all().map((c) => c.name);
  if (!cols.includes('claimed_by'))    db.exec(`ALTER TABLE jobs ADD COLUMN claimed_by TEXT NOT NULL DEFAULT ''`);
  if (!cols.includes('claimed_at'))    db.exec(`ALTER TABLE jobs ADD COLUMN claimed_at DATETIME`);
  if (!cols.includes('attempts'))      db.exec(`ALTER TABLE jobs ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0`);
  if (!cols.includes('max_attempts'))  db.exec(`ALTER TABLE jobs ADD COLUMN max_attempts INTEGER NOT NULL DEFAULT 3`);
  if (!cols.includes('next_retry_at')) db.exec(`ALTER TABLE jobs ADD COLUMN next_retry_at DATETIME`);
}

// ── Logging ───────────────────────────────────────────────────────────────────

function log(level, message, context = {}) {
  const line = `${new Date().toISOString()} [${level}] ${message} ${JSON.stringify(context)}`;
  try { fs.appendFileSync(LOG_PATH, `${line}\n`); } catch { /* no-op */ }
  if (level === 'ERROR')      console.error(line);
  else if (level === 'WARN')  console.warn(line);
  else                        console.log(line);
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
