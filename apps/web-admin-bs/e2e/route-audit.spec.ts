import { test, expect, type Page } from '@playwright/test';
import * as fs from 'fs';

const BASE = process.env.E2E_BASE_URL || 'http://localhost:8222';
const ROUTES = [
  '/',
  '/pusaka',
  '/pusaka/employees',
  '/pusaka/kehadiran',
  '/pusaka/summary',
  '/pusaka/telegram-laporan',
  '/pusaka/antrian',
  '/employees',
  '/settings/account',
  '/settings/school-profile',
  '/settings/users',
  '/settings/rbac',
  '/settings/branding',
  '/settings/backups',
  '/settings/audit-logs',
  '/settings/analytics',
];

function getEnvPassword(): string {
  const envPath = '/home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api/.env';
  const content = fs.readFileSync(envPath, 'utf-8');
  for (const line of content.split('\n')) {
    if (line.includes('ADMIN') && line.includes('PASSWORD') && line.includes('=')) {
      return line.substring(line.indexOf('=') + 1).trim();
    }
  }
  throw new Error('ADMIN password not found');
}

async function login(page: Page) {
  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', getEnvPassword());
  await page.click('button[type="submit"]');
  await page.waitForFunction(() => !location.pathname.includes('/login'), null, { timeout: 15000 });
}

for (const route of ROUTES) {
  test(`route opens: ${route}`, async ({ page }, testInfo) => {
    const pageErrors: string[] = [];
    const failedRequests: string[] = [];
    const badResponses: string[] = [];

    page.on('pageerror', (err) => pageErrors.push(err.message));
    page.on('requestfailed', (req) => failedRequests.push(`${req.method()} ${req.url()} :: ${req.failure()?.errorText}`));
    page.on('response', (res) => {
      const status = res.status();
      const url = res.url();
      if (status >= 400 && !url.includes('/favicon')) {
        badResponses.push(`${status} ${url}`);
      }
    });

    await login(page);
    await page.goto(`${BASE}${route}`, { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(1500);

    const url = page.url();
    const text = await page.locator('body').innerText({ timeout: 5000 }).catch(() => '');
    const hasShell = await page.locator('.page-header').isVisible().catch(() => false);
    const hasSidebar = await page.locator('#sidebarOffcanvas, aside.sidebar').count();
    const errorLike = /Terjadi kesalahan|Sistem sedang mengalami kendala|\b(404|500)\b\s+(Not Found|Internal)|Not found|Gagal memuat/i.test(text);
    const stuckLoading = /Memuat/i.test(text) && text.length < 600;

    await testInfo.attach('route-audit', {
      body: JSON.stringify({ route, url, hasShell, hasSidebar, errorLike, stuckLoading, pageErrors, failedRequests, badResponses, text: text.slice(0, 1500) }, null, 2),
      contentType: 'application/json',
    });

    expect(url, `route ${route} redirected to login or wrong page`).not.toContain('/login');
    expect(hasShell, `route ${route} missing app shell`).toBe(true);
    expect(hasSidebar, `route ${route} missing sidebar/offcanvas`).toBeGreaterThan(0);
    expect(pageErrors, `route ${route} has JS errors`).toEqual([]);
    expect(errorLike, `route ${route} shows error text: ${text.slice(0, 300)}`).toBe(false);
    expect(stuckLoading, `route ${route} appears stuck loading`).toBe(false);
  });
}
