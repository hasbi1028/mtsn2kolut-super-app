import fs from 'fs';
import path from 'path';

import { chromium, type Page } from 'playwright';

import {
  ACTION_TIMEOUT,
  BASE_LAT,
  BASE_LNG,
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

export async function processClaimedJob(
  job: ClaimedJob,
  runtimeConfig: RuntimeConfig,
): Promise<AttendanceRecord> {
  if (job.run_type === 'checkin') {
    await withRetry(
      (attempt) =>
        checkin(job.pusaka_username, job.pusaka_password, attempt, runtimeConfig),
      'checkin',
    );
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
        ),
      'checkout',
    );
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
      ),
    'scrape',
  );
}

async function withRetry<T>(
  fn: (attempt: number) => Promise<T>,
  label: string,
): Promise<T> {
  let lastError: unknown;
  for (let attempt = 1; attempt <= SCRAPE_RETRIES; attempt++) {
    try {
      return await fn(attempt);
    } catch (error) {
      lastError = error;
      log('WARN', `${label} attempt ${attempt}/${SCRAPE_RETRIES} failed`, {
        error: (error as Error)?.message ?? String(error),
      });
      if (attempt < SCRAPE_RETRIES) {
        await sleep(SCRAPE_RETRY_MS * attempt);
      }
    }
  }
  throw lastError;
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

    log('INFO', `${label}: login ok`, { username, url: page.url() });
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
      username,
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
    log('INFO', `${label}: BERHASIL`, { username });
    return;
  }
  if (/sudah presensi/i.test(text)) {
    log('WARN', `${label}: sudah absen sebelumnya`, { username });
    return;
  }
  log('WARN', `${label}: status tidak jelas`, {
    username,
    snippet: text.substring(0, 300),
  });
}

async function saveFailScreenshot(page: Page, name: string): Promise<void> {
  try {
    fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const file = path.join(SCREENSHOT_DIR, `${name}-${timestamp}.png`);
    await page.screenshot({ path: file, fullPage: false });
    log('WARN', 'screenshot saved', { file });
  } catch {
    // best-effort
  }
}

async function scrapeOnce(
  username: string,
  password: string,
  attempt: number,
  runtimeConfig: RuntimeConfig,
): Promise<AttendanceRecord> {
  const browser = await chromium.launch(getBrowserOpts(runtimeConfig));
  const context = await browser.newContext(getContextOpts());
  context.setDefaultTimeout(ACTION_TIMEOUT);
  const page = await context.newPage();

  try {
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
    await page.waitForTimeout(2000);

    const parsed = parseTodayFromText(
      await page.locator('body').innerText(),
      getTodayLabelID(),
    );
    if (!parsed) {
      throw new Error(`Data hari ini tidak ditemukan: ${getTodayLabelID()}`);
    }
    return parsed;
  } catch (error) {
    await saveFailScreenshot(page, `scrape-fail-a${attempt}`);
    throw error;
  } finally {
    await context.close();
    await browser.close();
  }
}

async function checkin(
  username: string,
  password: string,
  attempt: number,
  runtimeConfig: RuntimeConfig,
): Promise<void> {
  const label = 'checkin';
  const geo = randomGeo(BASE_LAT, BASE_LNG, 28);
  const browser = await chromium.launch(getBrowserOpts(runtimeConfig));
  const context = await browser.newContext(
    getContextOpts({ geolocation: geo, permissions: ['geolocation'] }),
  );
  context.setDefaultTimeout(ACTION_TIMEOUT);
  const page = await context.newPage();

  try {
    log('INFO', `${label}: starting`, {
      username,
      lat: geo.latitude.toFixed(7),
      lng: geo.longitude.toFixed(7),
    });
    await loginToPusaka(page, username, password, label);
    await page.goto(`${BASE_URL}/profile/presence`, {
      waitUntil: 'domcontentloaded',
    });
    await page.waitForTimeout(3000);
    await triggerGeo(page);
    await page.waitForTimeout(2000);

    const button = page.locator('button:has-text("Presensi masuk")');
    if ((await button.count()) === 0) {
      if (
        /hadir|sudah/i.test(await page.locator('body').innerText().catch(() => ''))
      ) {
        log('WARN', `${label}: sudah absen`, { username });
        return;
      }
      throw new Error('Tombol Presensi masuk tidak ditemukan');
    }
    if (await button.first().isDisabled().catch(() => false)) {
      log('WARN', `${label}: tombol disabled`, { username });
      return;
    }

    await triggerGeo(page);
    await page.waitForTimeout(1000);
    await button.first().click();
    log('INFO', `${label}: tombol diklik`, { username });
    await page.waitForTimeout(3000);
    await assertPresenceResult(page, label, username);
  } catch (error) {
    await saveFailScreenshot(page, `${label}-fail-a${attempt}-${username}`);
    throw error;
  } finally {
    await context.close();
    await browser.close();
  }
}

async function checkout(
  username: string,
  password: string,
  attempt: number,
  runtimeConfig: RuntimeConfig,
): Promise<void> {
  const label = 'checkout';
  const geo = randomGeo(BASE_LAT, BASE_LNG, 28);
  const browser = await chromium.launch(getBrowserOpts(runtimeConfig));
  const context = await browser.newContext(
    getContextOpts({ geolocation: geo, permissions: ['geolocation'] }),
  );
  context.setDefaultTimeout(ACTION_TIMEOUT);
  const page = await context.newPage();

  try {
    log('INFO', `${label}: starting`, {
      username,
      lat: geo.latitude.toFixed(7),
      lng: geo.longitude.toFixed(7),
    });
    await loginToPusaka(page, username, password, label);
    await page.goto(`${BASE_URL}/profile/presence`, {
      waitUntil: 'domcontentloaded',
    });
    await page.waitForTimeout(3000);
    await triggerGeo(page);
    await page.waitForTimeout(2000);

    const button = page.locator('button:has-text("Presensi pulang")');
    if ((await button.count()) === 0) {
      throw new Error('Tombol Presensi pulang tidak ditemukan');
    }
    if (await button.first().isDisabled().catch(() => false)) {
      log('WARN', `${label}: tombol disabled`, { username });
      return;
    }

    await button.first().click();
    log('INFO', `${label}: tombol diklik`, { username });
    await page.waitForTimeout(2000);

    const yesButton = page.getByRole('button', { name: /^ya$/i });
    if (
      (await yesButton.count()) > 0 &&
      !(await yesButton.first().isDisabled().catch(() => false))
    ) {
      await yesButton.first().click();
      log('INFO', `${label}: konfirmasi Ya diklik`, { username });
    }

    await page.waitForTimeout(3000);
    await assertPresenceResult(page, label, username);
  } catch (error) {
    await saveFailScreenshot(page, `${label}-fail-a${attempt}-${username}`);
    throw error;
  } finally {
    await context.close();
    await browser.close();
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
