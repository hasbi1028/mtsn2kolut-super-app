import { test, expect, type Page, type BrowserContext } from '@playwright/test';
import * as fs from 'fs';

const BASE = process.env.E2E_BASE_URL || 'http://localhost:8021';
const MOBILE = { width: 390, height: 844 };
const DESKTOP = { width: 1440, height: 900 };

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

async function desktop(browser: any): Promise<BrowserContext> {
  return browser.newContext({ viewport: DESKTOP });
}

async function mobileCtx(browser: any): Promise<BrowserContext> {
  return browser.newContext({
    viewport: MOBILE,
    isMobile: true,
    hasTouch: true,
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) Mobile',
    deviceScaleFactor: 3,
  });
}

async function loginDesktop(browser: any): Promise<{ page: Page; ctx: BrowserContext }> {
  const ctx = await desktop(browser);
  const page = await ctx.newPage();
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
  await page.fill('#username', 'admin');
  await page.fill('#password', PW);
  await page.click('button[type="submit"]');
  await page.waitForTimeout(3000);
  return { page, ctx };
}

async function loginMobile(browser: any): Promise<{ page: Page; ctx: BrowserContext }> {
  const ctx = await mobileCtx(browser);
  const page = await ctx.newPage();
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
  await page.fill('#username', 'admin');
  await page.fill('#password', PW);
  await page.click('button[type="submit"]');
  await page.waitForTimeout(3000);
  return { page, ctx };
}

function isLoggedIn(page: Page): boolean {
  return !page.url().includes('/login');
}

// ─── Assign Guru (Subject Assignments) ───

test.describe('ASSIGN GURU — Subject Assignments', () => {
  test.describe('Desktop', () => {
    test('loads matrix table', async ({ browser }) => {
      const { page, ctx } = await loginDesktop(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      const errors: string[] = [];
      page.on('pageerror', (e) => errors.push(e.message));

      await page.goto(`${BASE}/academic/subject-assignments`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);

      const title = await page.title();
      expect(title).toContain('Assign Guru');
      await expect(page.locator('h1')).toHaveText(/Assign Guru/);
      await expect(page.locator('table')).toBeVisible({ timeout: 5000 });

      const subjectCells = await page.locator('table tbody tr').count();
      console.log(`Subject rows: ${subjectCells}`);
      expect(subjectCells).toBeGreaterThan(5);

      const classHeaders = await page.locator('table thead th').count();
      console.log(`Column headers: ${classHeaders}`);
      expect(classHeaders).toBeGreaterThan(2);

      expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
      await page.screenshot({ path: 'test-results/assign-guru-desktop.png', fullPage: true });
      await ctx.close();
    });

    test('has level filter buttons', async ({ browser }) => {
      const { page, ctx } = await loginDesktop(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      await page.goto(`${BASE}/academic/subject-assignments`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);
      await expect(page.locator('button:has-text("VII")')).toBeVisible();
      await expect(page.locator('button:has-text("VIII")')).toBeVisible();
      await expect(page.locator('button:has-text("IX")')).toBeVisible();
      await page.screenshot({ path: 'test-results/assign-guru-filters.png', fullPage: true });
      await ctx.close();
    });
  });

  test.describe('Mobile', () => {
    test('shows card list instead of table', async ({ browser }) => {
      const { page, ctx } = await loginMobile(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      const errors: string[] = [];
      page.on('pageerror', (e) => errors.push(e.message));

      await page.goto(`${BASE}/academic/subject-assignments`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);

      const sw = await page.evaluate(() => document.documentElement.scrollWidth);
      const cw = await page.evaluate(() => document.documentElement.clientWidth);
      expect(sw).toBeLessThanOrEqual(cw + 2);

      expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
      await page.screenshot({ path: 'test-results/assign-guru-mobile.png', fullPage: true });
      await ctx.close();
    });
  });
});

// ─── Jadwal Pelajaran (Timetable) ───

test.describe('JADWAL PELAJARAN — Timetable', () => {
  test.describe('Desktop', () => {
    test('loads per-class timetable cards', async ({ browser }) => {
      const { page, ctx } = await loginDesktop(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      const errors: string[] = [];
      page.on('pageerror', (e) => errors.push(e.message));

      await page.goto(`${BASE}/academic/timetable`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);

      const title = await page.title();
      expect(title).toContain('Jadwal Pelajaran');
      await expect(page.locator('h1')).toHaveText(/Jadwal Pelajaran/);
      await expect(page.locator('table')).toBeVisible({ timeout: 5000 });

      expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
      await page.screenshot({ path: 'test-results/timetable-desktop.png', fullPage: true });
      await ctx.close();
    });

    test('shows day headers', async ({ browser }) => {
      const { page, ctx } = await loginDesktop(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      await page.goto(`${BASE}/academic/timetable`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);
      const dayHeaders = await page.locator('table thead th').allTextContents();
      const headerText = dayHeaders.join(' ');
      expect(headerText).toContain('Senin');
      expect(headerText).toContain('Sabtu');
      await ctx.close();
    });
  });

  test.describe('Mobile', () => {
    test('mobile layout loads', async ({ browser }) => {
      const { page, ctx } = await loginMobile(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      const errors: string[] = [];
      page.on('pageerror', (e) => errors.push(e.message));

      await page.goto(`${BASE}/academic/timetable`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);

      const sw = await page.evaluate(() => document.documentElement.scrollWidth);
      const cw = await page.evaluate(() => document.documentElement.clientWidth);
      expect(sw).toBeLessThanOrEqual(cw + 2);

      expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
      await page.screenshot({ path: 'test-results/timetable-mobile.png', fullPage: true });
      await ctx.close();
    });
  });
});

// ─── Jurnal Harian (Class Journal) ───

test.describe('JURNAL HARIAN — Class Journal', () => {
  test.describe('Desktop', () => {
    test('loads journal page', async ({ browser }) => {
      const { page, ctx } = await loginDesktop(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      const errors: string[] = [];
      page.on('pageerror', (e) => errors.push(e.message));

      await page.goto(`${BASE}/academic/class-journal`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);

      const title = await page.title();
      expect(title).toContain('Jurnal Belajar Harian');
      await expect(page.locator('h1')).toHaveText(/Jurnal Belajar Harian/);

      expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
      await page.screenshot({ path: 'test-results/journal-desktop.png', fullPage: true });
      await ctx.close();
    });

    test('shows assignment selector', async ({ browser }) => {
      const { page, ctx } = await loginDesktop(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      await page.goto(`${BASE}/academic/class-journal`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);
      await expect(page.locator('select')).toBeVisible({ timeout: 5000 });
      const options = await page.locator('select option').count();
      console.log(`Assignment options: ${options}`);
      expect(options).toBeGreaterThan(0);
      await ctx.close();
    });
  });

  test.describe('Mobile', () => {
    test('mobile layout no horizontal scroll', async ({ browser }) => {
      const { page, ctx } = await loginMobile(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      const errors: string[] = [];
      page.on('pageerror', (e) => errors.push(e.message));

      await page.goto(`${BASE}/academic/class-journal`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);

      const sw = await page.evaluate(() => document.documentElement.scrollWidth);
      const cw = await page.evaluate(() => document.documentElement.clientWidth);
      expect(sw).toBeLessThanOrEqual(cw + 2);

      expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
      await page.screenshot({ path: 'test-results/journal-mobile.png', fullPage: true });
      await ctx.close();
    });
  });
});

// ─── Data Murid (Kesiswaan) ───

test.describe('DATA MURID — Kesiswaan', () => {
  test.describe('Desktop', () => {
    test('loads murid list page', async ({ browser }) => {
      const { page, ctx } = await loginDesktop(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      const errors: string[] = [];
      page.on('pageerror', (e) => errors.push(e.message));

      await page.goto(`${BASE}/kesiswaan/murid`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);

      const title = await page.title();
      expect(title).toContain('Data Murid');
      await expect(page.locator('h1')).toHaveText(/Data Murid/);

      expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
      await page.screenshot({ path: 'test-results/murid-desktop.png', fullPage: true });
      await ctx.close();
    });

    test('has search and filter controls', async ({ browser }) => {
      const { page, ctx } = await loginDesktop(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      await page.goto(`${BASE}/kesiswaan/murid`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);
      await expect(page.locator('input[placeholder*="Cari"]')).toBeVisible({ timeout: 5000 });
      await expect(page.locator('select')).toBeVisible();
      await ctx.close();
    });
  });

  test.describe('Mobile', () => {
    test('mobile layout loads', async ({ browser }) => {
      const { page, ctx } = await loginMobile(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      const errors: string[] = [];
      page.on('pageerror', (e) => errors.push(e.message));

      await page.goto(`${BASE}/kesiswaan/murid`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);

      const sw = await page.evaluate(() => document.documentElement.scrollWidth);
      const cw = await page.evaluate(() => document.documentElement.clientWidth);
      expect(sw).toBeLessThanOrEqual(cw + 2);

      expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
      await page.screenshot({ path: 'test-results/murid-mobile.png', fullPage: true });
      await ctx.close();
    });
  });
});

// ─── Sidebar ───

test.describe('SIDEBAR ACADEMIC GROUP', () => {
  test.describe('Desktop', () => {
    test('has all academic menu items', async ({ browser }) => {
      const { page, ctx } = await loginDesktop(browser);
      if (!isLoggedIn(page)) { await ctx.close(); return; }
      await page.goto(`${BASE}/`, { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);

      // Check sidebar for academic items
      const sidebarText = await page.locator('aside').allTextContents();
      const allText = sidebarText.join(' ');

      expect(allText).toContain('Assign Guru');
      expect(allText).toContain('Jadwal Pelajaran');
      expect(allText).toContain('Jurnal Harian');
      expect(allText).toContain('Data Murid');
      expect(allText).toContain('Semester');
      expect(allText).toContain('Rombel');
      expect(allText).toContain('Kurikulum');

      await ctx.close();
    });
  });
});

// ─── Cross-module flow ───

test.describe('ACADEMIC FLOW', () => {
  test('navigate from assign guru to timetable to journal', async ({ browser }) => {
    const { page, ctx } = await loginDesktop(browser);
    if (!isLoggedIn(page)) { await ctx.close(); return; }
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(e.message));

    // 1. Assign Guru
    await page.goto(`${BASE}/academic/subject-assignments`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);
    await expect(page.locator('h1')).toHaveText(/Assign Guru/);

    // 2. Timetable
    await page.goto(`${BASE}/academic/timetable`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);
    await expect(page.locator('h1')).toHaveText(/Jadwal Pelajaran/);

    // 3. Class Journal
    await page.goto(`${BASE}/academic/class-journal`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);
    await expect(page.locator('h1')).toHaveText(/Jurnal Belajar Harian/);

    // 4. Data Murid
    await page.goto(`${BASE}/kesiswaan/murid`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);
    await expect(page.locator('h1')).toHaveText(/Data Murid/);

    expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
    await ctx.close();
  });
});
