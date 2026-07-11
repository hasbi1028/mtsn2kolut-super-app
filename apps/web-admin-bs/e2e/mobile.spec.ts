import { test, expect, type Page, type BrowserContext } from '@playwright/test';
import * as fs from 'fs';

const BASE = 'http://localhost:8222';
const MOBILE = { width: 390, height: 844 };

function getEnvPassword(): string {
  const envPath = '/home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api/.env';
  const content = fs.readFileSync(envPath, 'utf-8');
  const lines = content.split('\n');
  for (const l of lines) {
    if (l.includes('ADMIN') && l.includes('PASSWORD') && l.includes('=')) {
      return l.substring(l.indexOf('=') + 1);
    }
  }
  return '';
}

const PW = getEnvPassword();

async function fresh(browser: any): Promise<BrowserContext> {
  return browser.newContext({
    viewport: MOBILE,
    isMobile: true,
    hasTouch: true,
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) Mobile',
    deviceScaleFactor: 3,
  });
}

async function login(browser: any): Promise<{ page: Page; ctx: BrowserContext }> {
  const ctx = await fresh(browser);
  const page = await ctx.newPage();
  await page.goto(BASE + '/login', { waitUntil: 'networkidle' });
  await page.fill('#username', 'admin');
  await page.fill('#password', PW);
  await page.click('button[type="submit"]');
  await page.waitForTimeout(3000);
  return { page, ctx };
}

function ok(page: Page): boolean {
  return !page.url().includes('/login');
}

test.describe('LOGIN PAGE MOBILE', () => {
  test('loads on mobile', async ({ browser }) => {
    const ctx = await fresh(browser);
    const page = await ctx.newPage();
    await page.goto(BASE, { waitUntil: 'networkidle' });
    await expect(page).toHaveURL(/login/);
    await expect(page.locator('.login-card')).toBeVisible();
    await ctx.close();
  });

  test('no horizontal scroll', async ({ browser }) => {
    const ctx = await fresh(browser);
    const page = await ctx.newPage();
    await page.goto(BASE + '/login', { waitUntil: 'networkidle' });
    const sw = await page.evaluate(() => document.documentElement.scrollWidth);
    const cw = await page.evaluate(() => document.documentElement.clientWidth);
    expect(sw).toBeLessThanOrEqual(cw + 2);
    await ctx.close();
  });

  test('inputs touch-friendly', async ({ browser }) => {
    const ctx = await fresh(browser);
    const page = await ctx.newPage();
    await page.goto(BASE + '/login', { waitUntil: 'networkidle' });
    const ub = await page.locator('#username').boundingBox();
    expect(ub!.height).toBeGreaterThanOrEqual(36);
    const pb = await page.locator('#password').boundingBox();
    expect(pb!.height).toBeGreaterThanOrEqual(36);
    const sb = await page.locator('button[type="submit"]').boundingBox();
    expect(sb!.height).toBeGreaterThanOrEqual(44);
    await ctx.close();
  });

  test('password toggle', async ({ browser }) => {
    const ctx = await fresh(browser);
    const page = await ctx.newPage();
    await page.goto(BASE + '/login', { waitUntil: 'networkidle' });
    await expect(page.locator('#password')).toHaveAttribute('type', 'password');
    await page.click('button[aria-label="Toggle password visibility"]');
    await expect(page.locator('#password')).toHaveAttribute('type', 'text');
    await page.click('button[aria-label="Toggle password visibility"]');
    await expect(page.locator('#password')).toHaveAttribute('type', 'password');
    await ctx.close();
  });

  test('wrong credentials shows error', async ({ browser }) => {
    const ctx = await fresh(browser);
    const page = await ctx.newPage();
    await page.goto(BASE + '/login', { waitUntil: 'networkidle' });
    await page.fill('#username', 'wronguser');
    await page.fill('#password', 'wrongpass');
    await page.click('button[type="submit"]');
    await expect(page.locator('.alert-danger')).toBeVisible({ timeout: 10000 });
    await ctx.close();
  });

  test('valid login redirects to dashboard', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    expect(ok(page)).toBe(true);
    await ctx.close();
  });
});

test.describe('DASHBOARD MOBILE LAYOUT', () => {
  test('desktop sidebar hidden', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    if (ok(page)) {
      await expect(page.locator('aside.sidebar')).not.toBeVisible();
    }
    await ctx.close();
  });

  test('hamburger visible', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    if (ok(page)) {
      await expect(page.locator('button[data-bs-target="#sidebarOffcanvas"]')).toBeVisible();
    }
    await ctx.close();
  });

  test('page header visible', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    if (ok(page)) {
      await expect(page.locator('.page-header')).toBeVisible();
    }
    await ctx.close();
  });

  test('no horizontal scroll', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    if (ok(page)) {
      const sw = await page.evaluate(() => document.documentElement.scrollWidth);
      const cw = await page.evaluate(() => document.documentElement.clientWidth);
      expect(sw).toBeLessThanOrEqual(cw + 2);
    }
    await ctx.close();
  });
});

test.describe('MOBILE SIDEBAR', () => {
  test('opens on hamburger click', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    if (ok(page)) {
      await page.click('button[data-bs-target="#sidebarOffcanvas"]');
      await page.waitForTimeout(600);
      await expect(page.locator('#sidebarOffcanvas')).toHaveClass(/show/);
    }
    await ctx.close();
  });

  test('has all nav items', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    if (ok(page)) {
      await page.click('button[data-bs-target="#sidebarOffcanvas"]');
      await page.waitForTimeout(600);
      const count = await page.locator('#sidebarOffcanvas .nav-link').count();
      expect(count).toBeGreaterThan(5);
    }
    await ctx.close();
  });

  test('closes after nav click', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    if (ok(page)) {
      await page.click('button[data-bs-target="#sidebarOffcanvas"]');
      await page.waitForTimeout(600);
      await page.locator('#sidebarOffcanvas .nav-link').first().click();
      await page.waitForTimeout(600);
      const hasShow = await page.locator('#sidebarOffcanvas').evaluate(
        (el: Element) => el.classList.contains('show')
      );
      expect(hasShow).toBe(false);
    }
    await ctx.close();
  });

  test('closes via X button', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    if (ok(page)) {
      await page.click('button[data-bs-target="#sidebarOffcanvas"]');
      await page.waitForTimeout(600);
      await page.click('#sidebarOffcanvas .btn-close');
      await page.waitForTimeout(600);
      const hasShow = await page.locator('#sidebarOffcanvas').evaluate(
        (el: Element) => el.classList.contains('show')
      );
      expect(hasShow).toBe(false);
    }
    await ctx.close();
  });
});

test.describe('SCREENSHOTS', () => {
  test('login page', async ({ browser }) => {
    const ctx = await fresh(browser);
    const page = await ctx.newPage();
    await page.goto(BASE + '/login', { waitUntil: 'networkidle' });
    await page.screenshot({ path: 'test-results/mobile-login.png', fullPage: true });
    await ctx.close();
  });

  test('dashboard', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    if (ok(page)) {
      await page.screenshot({ path: 'test-results/mobile-dashboard.png', fullPage: true });
    }
    await ctx.close();
  });

  test('sidebar open', async ({ browser }) => {
    const { page, ctx } = await login(browser);
    if (ok(page)) {
      await page.click('button[data-bs-target="#sidebarOffcanvas"]');
      await page.waitForTimeout(600);
      await page.screenshot({ path: 'test-results/mobile-sidebar.png', fullPage: true });
    }
    await ctx.close();
  });
});

