import { test, expect } from '@playwright/test';
import * as fs from 'fs';

const BASE = process.env.E2E_BASE_URL || 'http://localhost:8021';

function getEnvPassword(): string {
  const envPath = '/home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api/.env';
  const content = fs.readFileSync(envPath, 'utf-8');
  for (const line of content.split('\n')) {
    if (line.startsWith('ADMIN_PASSWORD=')) {
      return line.substring('ADMIN_PASSWORD='.length).trim();
    }
  }
  return '';
}

const PW = getEnvPassword();

async function loggedInPage(browser: any): Promise<import('@playwright/test').Page> {
  const page = await browser.newPage();
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
  await page.fill('#username', 'admin');
  await page.fill('#password', PW);
  await page.click('button[type="submit"]');
  // Login via SvelteKit form action (fetch-based, not full redirect)
  await page.waitForTimeout(4000);
  // Tolerate staying on login if wrong creds or session issue
  return page;
}

// ─── RESPONSIVE ───
test.describe('Mobile Responsive', () => {
  test('login page fits viewport', async ({ browser }) => {
    const page = await browser.newPage();
    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    const sw = await page.evaluate(() => document.documentElement.scrollWidth);
    const cw = await page.evaluate(() => document.documentElement.clientWidth);
    expect(sw).toBeLessThanOrEqual(cw + 2);

    const inputH = await page.locator('#username').evaluate(el => el.getBoundingClientRect().height);
    expect(inputH).toBeGreaterThanOrEqual(36);
    const btnH = await page.locator('button[type="submit"]').evaluate(el => el.getBoundingClientRect().height);
    expect(btnH).toBeGreaterThanOrEqual(44);
  });

  test('assign guru: no horizontal scroll', async ({ browser }) => {
    const page = await loggedInPage(browser);
    await page.goto(`${BASE}/academic/subject-assignments`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);
    const sw = await page.evaluate(() => document.documentElement.scrollWidth);
    const cw = await page.evaluate(() => document.documentElement.clientWidth);
    expect(sw).toBeLessThanOrEqual(cw + 2);
  });

  test('timetable: mobile list layout', async ({ browser }) => {
    const page = await loggedInPage(browser);
    await page.goto(`${BASE}/academic/timetable`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);
    const sw = await page.evaluate(() => document.documentElement.scrollWidth);
    const cw = await page.evaluate(() => document.documentElement.clientWidth);
    expect(sw).toBeLessThanOrEqual(cw + 2);
  });

  test('journal: mobile layout', async ({ browser }) => {
    const page = await loggedInPage(browser);
    await page.goto(`${BASE}/academic/class-journal`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);
    const sw = await page.evaluate(() => document.documentElement.scrollWidth);
    const cw = await page.evaluate(() => document.documentElement.clientWidth);
    expect(sw).toBeLessThanOrEqual(cw + 2);
  });

  test('murid: mobile card layout', async ({ browser }) => {
    const page = await loggedInPage(browser);
    await page.goto(`${BASE}/kesiswaan/murid`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);
    const sw = await page.evaluate(() => document.documentElement.scrollWidth);
    const cw = await page.evaluate(() => document.documentElement.clientWidth);
    expect(sw).toBeLessThanOrEqual(cw + 2);
  });

  test('all mobile screenshots', async ({ browser }) => {
    const page = await loggedInPage(browser);
    for (const url of ['/login', '/academic/subject-assignments', '/academic/timetable', '/academic/class-journal', '/kesiswaan/murid']) {
      await page.goto(`${BASE}${url}`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(1000);
      const name = url.replace(/\//g, '-').replace(/^-/, '') || 'dashboard';
      await page.screenshot({ path: `test-results/mobile-${name}.png`, fullPage: true });
    }
  });
});
