import fs from 'fs';
import crypto from 'crypto';
import path from 'path';

import { chromium, type Browser, type BrowserContext, type Page } from 'playwright';

import {
  ACTION_TIMEOUT,
  BASE_URL,
  BROWSER_ARGS,
  SCRAPE_RETRIES,
  SCRAPE_RETRY_MS,
  SCREENSHOT_DIR,
  USER_AGENT,
} from './config.js';
import { log } from './logger.js';
import {
  getTodayLabelID,
  parseTodayFromText,
  randomGeo,
  toISODateMakassar,
  toTimeWITA,
} from './parsers.js';
import type { AttendanceRecord, ClaimedJob, RuntimeConfig } from './types.js';

const SCREENSHOT_RETENTION_MS = 7 * 24 * 60 * 60 * 1000;

function usernameHash(username: string): string {
  return crypto.createHash('sha256').update(username).digest('hex').slice(0, 12);
}

function usernameContext(username: string): Record<string, string> {
  return { username_hash: usernameHash(username) };
}

function pageTextSummary(text: string): Record<string, unknown> {
  const normalized = text.replace(/\s+/g, ' ').trim();
  const markers = [
    ['login_form', /username|password|masuk/i],
    ['attendance', /absensi|presensi|hadir/i],
    ['dashboard', /dashboard|beranda|profil/i],
    ['error', /gagal|error|bad request|timeout/i],
  ]
    .filter(([, pattern]) => (pattern as RegExp).test(normalized))
    .map(([marker]) => marker);

  return {
    text_hash: crypto.createHash('sha256').update(normalized).digest('hex').slice(0, 12),
    text_length: normalized.length,
    markers,
  };
}

function sanitizeScreenshotName(name: string): string {
  return name.replace(/[^a-z0-9._-]+/gi, '-').slice(0, 120);
}

function isCredentialOrLoginFailure(error: unknown): boolean {
  const message = ((error as Error | undefined)?.message ?? String(error)).toLowerCase();
  return /login|credential|username|password/.test(message);
}

export function pruneOldScreenshots(): void {
  if (!fs.existsSync(SCREENSHOT_DIR)) {
    return;
  }
  const cutoff = Date.now() - SCREENSHOT_RETENTION_MS;
  for (const entry of fs.readdirSync(SCREENSHOT_DIR, { withFileTypes: true })) {
    if (!entry.isFile()) {
      continue;
    }
    const file = path.join(SCREENSHOT_DIR, entry.name);
    const stat = fs.statSync(file);
    if (stat.mtimeMs < cutoff) {
      fs.unlinkSync(file);
    }
  }
}

export async function processClaimedJob(
  job: ClaimedJob,
  runtimeConfig: RuntimeConfig,
  signal?: AbortSignal,
): Promise<AttendanceRecord> {
  throwIfCancelled(signal);
  if (job.run_type === 'checkin') {
    await withRetry(
      (attempt) =>
        checkin(job.pusaka_username, job.pusaka_password, attempt, runtimeConfig, signal),
      'checkin',
      signal,
    );
    throwIfCancelled(signal);
    return {
      tanggal: toISODateMakassar(),
      jam_masuk: toTimeWITA(),
      jam_pulang: '',
    };
  }

  if (job.run_type === 'checkout') {
    await withRetry(
      (attempt) =>
        checkout(
          job.pusaka_username,
          job.pusaka_password,
          attempt,
          runtimeConfig,
          signal,
        ),
      'checkout',
      signal,
    );
    throwIfCancelled(signal);
    return {
      tanggal: toISODateMakassar(),
      jam_masuk: '',
      jam_pulang: toTimeWITA(),
    };
  }

  return withRetry<AttendanceRecord>(
    (attempt) =>
      scrapeOnce(
        job.pusaka_username,
        job.pusaka_password,
        attempt,
        runtimeConfig,
        signal,
      ),
    'scrape',
    signal,
  );
}

async function withRetry<T>(
  fn: (attempt: number) => Promise<T>,
  label: string,
  signal?: AbortSignal,
): Promise<T> {
  let lastError: unknown;
  for (let attempt = 1; attempt <= SCRAPE_RETRIES; attempt++) {
    throwIfCancelled(signal);
    try {
      return await fn(attempt);
    } catch (error) {
      lastError = error;
      log('WARN', `${label} attempt ${attempt}/${SCRAPE_RETRIES} failed`, {
        error: (error as Error)?.message ?? String(error),
      });
      if (attempt < SCRAPE_RETRIES) {
        await sleep(SCRAPE_RETRY_MS * attempt, signal);
      }
    }
  }
  throw lastError;
}

class WorkerJobCancelledError extends Error {
  constructor() {
    super('worker shutdown interrupted active job');
    this.name = 'WorkerJobCancelledError';
  }
}

function throwIfCancelled(signal?: AbortSignal): void {
  if (signal?.aborted) {
    throw new WorkerJobCancelledError();
  }
}

function registerAbortCleanup(
  signal: AbortSignal | undefined,
  resources: () => { page?: Page; context?: BrowserContext; browser?: Browser },
): () => void {
  if (!signal) {
    return () => {};
  }

  const closeResources = async () => {
    const { page, context, browser } = resources();
    await page?.close().catch(() => {});
    await context?.close().catch(() => {});
    await browser?.close().catch(() => {});
  };
  const onAbort = () => {
    void closeResources();
  };

  if (signal.aborted) {
    onAbort();
  } else {
    signal.addEventListener('abort', onAbort, { once: true });
  }

  return () => signal.removeEventListener('abort', onAbort);
}

function getBrowserOpts(runtimeConfig: RuntimeConfig) {
  return { headless: runtimeConfig.headless, args: BROWSER_ARGS };
}

function getContextOpts(extra: Record<string, unknown> = {}) {
  return {
    timezoneId: 'Asia/Makassar',
    locale: 'id-ID',
    viewport: { width: 1280, height: 800 },
    userAgent: USER_AGENT,
    ...extra,
  };
}

async function blockAssets(page: Page): Promise<void> {
  await page.route('**/*', (route) => {
    if (
      ['image', 'font', 'media', 'stylesheet'].includes(
        route.request().resourceType(),
      )
    ) {
      return route.abort();
    }
    route.continue();
  });
}

async function loginToPusaka(
  page: Page,
  username: string,
  password: string,
  label: string,
): Promise<void> {
  await page.goto(BASE_URL, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.getByRole('link', { name: /login/i }).first().waitFor({
    state: 'visible',
  });
  await page.getByRole('link', { name: /login/i }).first().click();
  await page.getByPlaceholder('Username').fill(username);
  await page.getByPlaceholder('Password').fill(password);

  const urlBeforeSubmit = page.url();
  await page.getByRole('button', { name: 'Masuk', exact: true }).click();

  try {
    await page.waitForURL(
      (url) => url.href !== urlBeforeSubmit && !/login/i.test(url.pathname),
      { timeout: 25000 },
    );

    const bodyAfter = await page.locator('body').innerText().catch(() => '');
    if (
      /username atau password salah|invalid credentials|login gagal/i.test(
        bodyAfter,
      )
    ) {
      throw new Error('Login gagal: username/password ditolak oleh server');
    }

    log('INFO', `${label}: login ok`, {
      ...usernameContext(username),
      url: page.url(),
    });
    return;
  } catch (urlError) {
    if ((urlError as Error)?.message?.includes('username/password')) {
      throw urlError;
    }
  }

  const body = await page.locator('body').innerText().catch(() => '');

  if (/username atau password salah|invalid credentials|login gagal/i.test(body)) {
    throw new Error('Login gagal: username/password ditolak oleh server');
  }

  if (/absensi|dashboard|beranda|profil|selamat\s+datang/i.test(body)) {
    log('INFO', `${label}: login ok (via body fallback)`, {
      ...usernameContext(username),
      url: page.url(),
    });
    return;
  }

  throw new Error(
    `Login timeout: halaman tidak merespons setelah submit (URL: ${page.url()})`,
  );
}

async function triggerGeo(page: Page): Promise<void> {
  await page
    .evaluate(
      () =>
        new Promise<void>((resolve) => {
          navigator.geolocation.getCurrentPosition(
            () => resolve(),
            () => resolve(),
            { enableHighAccuracy: true, timeout: 5000 },
          );
        }),
    )
    .catch(() => {});
}

async function assertPresenceResult(
  page: Page,
  label: string,
  username: string,
): Promise<void> {
  const text = await page.locator('body').innerText().catch(() => '');
  if (text.includes('PRESENSI GAGAL')) {
    throw new Error('PRESENSI GAGAL — Bad Request (mungkin GPS tidak valid)');
  }
  if (/berhasil/i.test(text)) {
    log('INFO', `${label}: BERHASIL`, usernameContext(username));
    return;
  }
  if (/sudah presensi/i.test(text)) {
    log('WARN', `${label}: sudah absen sebelumnya`, usernameContext(username));
    return;
  }
  log('WARN', `${label}: status tidak jelas`, {
    ...usernameContext(username),
    page: pageTextSummary(text),
  });
}

async function saveFailScreenshot(page: Page, name: string): Promise<void> {
  try {
    fs.mkdirSync(SCREENSHOT_DIR, { recursive: true, mode: 0o700 });
    fs.chmodSync(SCREENSHOT_DIR, 0o700);
    pruneOldScreenshots();
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const file = path.join(
      SCREENSHOT_DIR,
      `${sanitizeScreenshotName(name)}-${timestamp}.png`,
    );
    await page.screenshot({ path: file, fullPage: false });
    fs.chmodSync(file, 0o600);
    log('WARN', 'screenshot saved', { file });
  } catch {
    // best-effort
  }
}

async function saveOperationalFailScreenshot(
  page: Page,
  name: string,
  error: unknown,
): Promise<void> {
  if (isCredentialOrLoginFailure(error)) {
    log('WARN', 'screenshot skipped for credential/login failure');
    return;
  }

  await saveFailScreenshot(page, name);
}

async function scrapeOnce(
  username: string,
  password: string,
  attempt: number,
  runtimeConfig: RuntimeConfig,
  signal?: AbortSignal,
): Promise<AttendanceRecord> {
  throwIfCancelled(signal);
  const browser = await chromium.launch(getBrowserOpts(runtimeConfig));
  const context = await browser.newContext(getContextOpts());
  context.setDefaultTimeout(ACTION_TIMEOUT);
  const page = await context.newPage();
  const unregisterAbortCleanup = registerAbortCleanup(signal, () => ({
    page,
    context,
    browser,
  }));

  try {
    throwIfCancelled(signal);
    await loginToPusaka(page, username, password, 'scrape');
    await blockAssets(page);

    const absensiLink = page.getByRole('link', { name: /Absensi/i }).first();
    await absensiLink.waitFor({ state: 'visible' });
    await absensiLink.click();

    const riwayatButton = page.getByRole('button', {
      name: 'Riwayat Presensi',
      exact: true,
    });
    await riwayatButton.waitFor({ state: 'visible' });
    await riwayatButton.click();

    await page
      .getByText(/Jam Masuk|Sedang mengambil riwayat/i)
      .first()
      .waitFor()
      .catch(() => {});
    await sleep(2000, signal);
    throwIfCancelled(signal);

    const parsed = parseTodayFromText(
      await page.locator('body').innerText(),
      getTodayLabelID(),
    );
    if (!parsed) {
      throw new Error(`Data hari ini tidak ditemukan: ${getTodayLabelID()}`);
    }
    return parsed;
  } catch (error) {
    await saveOperationalFailScreenshot(page, `scrape-fail-a${attempt}`, error);
    throw error;
  } finally {
    unregisterAbortCleanup();
    await context.close().catch(() => {});
    await browser.close().catch(() => {});
  }
}

async function checkin(
  username: string,
  password: string,
  attempt: number,
  runtimeConfig: RuntimeConfig,
  signal?: AbortSignal,
): Promise<void> {
  throwIfCancelled(signal);
  const label = 'checkin';
  const geo = randomGeo(
    runtimeConfig.geo.baseLat,
    runtimeConfig.geo.baseLng,
    runtimeConfig.geo.checkinRadiusMeters,
  );
  const browser = await chromium.launch(getBrowserOpts(runtimeConfig));
  const context = await browser.newContext(
    getContextOpts({ geolocation: geo, permissions: ['geolocation'] }),
  );
  context.setDefaultTimeout(ACTION_TIMEOUT);
  const page = await context.newPage();
  const unregisterAbortCleanup = registerAbortCleanup(signal, () => ({
    page,
    context,
    browser,
  }));

  try {
    log('INFO', `${label}: starting`, {
      ...usernameContext(username),
      lat: geo.latitude.toFixed(7),
      lng: geo.longitude.toFixed(7),
    });
    throwIfCancelled(signal);
    await loginToPusaka(page, username, password, label);
    await page.goto(`${BASE_URL}/profile/presence`, {
      waitUntil: 'domcontentloaded',
    });
    await sleep(3000, signal);
    await triggerGeo(page);
    await sleep(2000, signal);

    const button = page.locator('button:has-text("Presensi masuk")');
    if ((await button.count()) === 0) {
      if (
        /hadir|sudah/i.test(await page.locator('body').innerText().catch(() => ''))
      ) {
        log('WARN', `${label}: sudah absen`, usernameContext(username));
        return;
      }
      throw new Error('Tombol Presensi masuk tidak ditemukan');
    }
    if (await button.first().isDisabled().catch(() => false)) {
      log('WARN', `${label}: tombol disabled`, usernameContext(username));
      return;
    }

    await triggerGeo(page);
    await sleep(1000, signal);
    throwIfCancelled(signal);
    await button.first().click();
    log('INFO', `${label}: tombol diklik`, usernameContext(username));
    await sleep(3000, signal);
    throwIfCancelled(signal);
    await assertPresenceResult(page, label, username);
  } catch (error) {
    await saveOperationalFailScreenshot(
      page,
      `${label}-fail-a${attempt}-u${usernameHash(username)}`,
      error,
    );
    throw error;
  } finally {
    unregisterAbortCleanup();
    await context.close().catch(() => {});
    await browser.close().catch(() => {});
  }
}

async function checkout(
  username: string,
  password: string,
  attempt: number,
  runtimeConfig: RuntimeConfig,
  signal?: AbortSignal,
): Promise<void> {
  throwIfCancelled(signal);
  const label = 'checkout';
  const geo = randomGeo(
    runtimeConfig.geo.baseLat,
    runtimeConfig.geo.baseLng,
    runtimeConfig.geo.checkoutRadiusMeters,
  );
  const browser = await chromium.launch(getBrowserOpts(runtimeConfig));
  const context = await browser.newContext(
    getContextOpts({ geolocation: geo, permissions: ['geolocation'] }),
  );
  context.setDefaultTimeout(ACTION_TIMEOUT);
  const page = await context.newPage();
  const unregisterAbortCleanup = registerAbortCleanup(signal, () => ({
    page,
    context,
    browser,
  }));

  try {
    log('INFO', `${label}: starting`, {
      ...usernameContext(username),
      lat: geo.latitude.toFixed(7),
      lng: geo.longitude.toFixed(7),
    });
    throwIfCancelled(signal);
    await loginToPusaka(page, username, password, label);
    await page.goto(`${BASE_URL}/profile/presence`, {
      waitUntil: 'domcontentloaded',
    });
    await sleep(3000, signal);
    await triggerGeo(page);
    await sleep(2000, signal);

    const button = page.locator('button:has-text("Presensi pulang")');
    if ((await button.count()) === 0) {
      throw new Error('Tombol Presensi pulang tidak ditemukan');
    }
    if (await button.first().isDisabled().catch(() => false)) {
      log('WARN', `${label}: tombol disabled`, usernameContext(username));
      return;
    }

    throwIfCancelled(signal);
    await button.first().click();
    log('INFO', `${label}: tombol diklik`, usernameContext(username));
    await sleep(2000, signal);

    const yesButton = page.getByRole('button', { name: /^ya$/i });
    if (
      (await yesButton.count()) > 0 &&
      !(await yesButton.first().isDisabled().catch(() => false))
    ) {
      await yesButton.first().click();
      log('INFO', `${label}: konfirmasi Ya diklik`, usernameContext(username));
    }

    await sleep(3000, signal);
    throwIfCancelled(signal);
    await assertPresenceResult(page, label, username);
  } catch (error) {
    await saveOperationalFailScreenshot(
      page,
      `${label}-fail-a${attempt}-u${usernameHash(username)}`,
      error,
    );
    throw error;
  } finally {
    unregisterAbortCleanup();
    await context.close().catch(() => {});
    await browser.close().catch(() => {});
  }
}

function sleep(ms: number, signal?: AbortSignal): Promise<void> {
  if (!signal) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  return new Promise((resolve, reject) => {
    const cleanup = () => signal.removeEventListener('abort', onAbort);
    const timeout = setTimeout(() => {
      cleanup();
      resolve();
    }, ms);
    const onAbort = () => {
      clearTimeout(timeout);
      cleanup();
      reject(new WorkerJobCancelledError());
    };

    if (signal.aborted) {
      onAbort();
      return;
    }

    signal.addEventListener('abort', onAbort, { once: true });
    timeout.unref?.();
  });
}
