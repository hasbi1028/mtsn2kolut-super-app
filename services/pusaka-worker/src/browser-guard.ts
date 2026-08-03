import { execFile } from 'node:child_process';
import fs from 'node:fs';
import process from 'node:process';
import { promisify } from 'node:util';

import { chromium } from 'playwright';

import { log } from './logger.js';

const execFileAsync = promisify(execFile);

/**
 * True jika executable Chromium Playwright tidak ditemukan.
 * Platform PaaS (mis. DomCloud) sering membersihkan cache browser
 * (~/.cache/ms-playwright) — ini deteksi awal untuk watchdog.
 */
export function browserExecutableMissing(): boolean {
  try {
    const executablePath = chromium.executablePath();
    return !fs.existsSync(executablePath);
  } catch {
    return true;
  }
}

/**
 * Pastikan Chromium ter-install. Kalau hilang, jalankan
 * `npx playwright install chromium` secara otomatis.
 *
 * Catatan: `PLAYWRIGHT_BROWSERS_PATH=0` dipakai supaya browser di-install
 * ke dalam folder project (node_modules/playwright-core/.local-browsers),
 * bukan ~/.cache — lebih tahan terhadap pembersihan cache platform.
 */
export async function ensureBrowserInstalled(): Promise<boolean> {
  if (!browserExecutableMissing()) {
    return true;
  }

  log('WARN', 'chromium executable missing — installing playwright chromium');
  try {
    const { stdout, stderr } = await execFileAsync(
      'npx',
      ['playwright', 'install', 'chromium'],
      {
        env: { ...process.env, PLAYWRIGHT_BROWSERS_PATH: '0' },
        timeout: 15 * 60 * 1000,
        maxBuffer: 16 * 1024 * 1024,
      },
    );
    log('INFO', 'chromium installed', { tail: (stdout + stderr).slice(-800) });
    return !browserExecutableMissing();
  } catch (error) {
    log('ERROR', 'chromium install failed', {
      error: (error as Error)?.message ?? String(error),
    });
    return false;
  }
}

/**
 * Watchdog periodik: jalankan ensureBrowserInstalled() tiap interval.
 * Return handle timer agar bisa di-unref (tidak menahan proses keluar).
 */
export function startBrowserWatchdog(intervalMs: number): NodeJS.Timeout {
  const timer = setInterval(() => {
    void ensureBrowserInstalled();
  }, intervalMs);
  timer.unref?.();
  return timer;
}
