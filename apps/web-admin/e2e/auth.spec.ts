import { test, expect, type Page } from '@playwright/test';
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

async function login(page: Page) {
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
  await page.fill('#username', 'admin');
  await page.fill('#password', PW);
  await page.click('button[type="submit"]');
  await page.waitForTimeout(3000);
}

test.describe('AUTH — Login Flow', () => {
  test('TC-AUTH-01: Landing page has working login button', async ({ browser }) => {
    const page = await browser.newPage();
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(e.message));

    await page.goto(`${BASE}/`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    // Should see the login button
    const loginLink = page.locator('a[href="/login"]').first();
    await expect(loginLink).toBeVisible({ timeout: 5000 });
    expect(await loginLink.textContent()).toContain('Masuk');

    // Click it → should navigate to /login
    await loginLink.click();
    await page.waitForTimeout(2000);
    expect(page.url()).toContain('/login');
    expect(errors).toEqual([]);
  });

  test('TC-AUTH-02: Login page renders with username, password, and submit', async ({ browser }) => {
    const page = await browser.newPage();
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(e.message));

    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1000);

    await expect(page).toHaveTitle(/Masuk/);
    await expect(page.locator('#username')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();

    // Check "Kembali ke Beranda" link exists
    const backLink = page.locator('a[href="/"]').first();
    await expect(backLink).toBeVisible();

    // Check toggle password button
    const toggleBtn = page.locator('button.toggle-password');
    await expect(toggleBtn).toBeVisible();
    expect(errors).toEqual([]);
  });

  test('TC-AUTH-03: Toggle password visibility works', async ({ browser }) => {
    const page = await browser.newPage();
    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1000);

    const pwField = page.locator('#password');
    expect(await pwField.getAttribute('type')).toBe('password');

    // Toggle show
    await page.click('button.toggle-password');
    await page.waitForTimeout(300);
    expect(await pwField.getAttribute('type')).toBe('text');
    expect(await page.locator('button.toggle-password').textContent()).toContain('Sembunyi');

    // Toggle hide back
    await page.click('button.toggle-password');
    await page.waitForTimeout(300);
    expect(await pwField.getAttribute('type')).toBe('password');
    expect(await page.locator('button.toggle-password').textContent()).toContain('Lihat');
  });

  test('TC-AUTH-04: Empty form does not submit (HTML5 validation)', async ({ browser }) => {
    const page = await browser.newPage();
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(e.message));

    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.click('button[type="submit"]');
    await page.waitForTimeout(1000);

    // Should stay on login page (HTML5 validation prevents submit)
    expect(page.url()).toContain('/login');
    expect(errors).toEqual([]);
  });

  test('TC-AUTH-05: Invalid credentials show error message', async ({ browser }) => {
    const page = await browser.newPage();
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(e.message));

    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'wrong_password_12345');
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);

    const body = await page.textContent('body') || '';
    expect(body.toLowerCase()).toContain('salah');
    // Should still be on login page
    expect(page.url()).toContain('/login');
    expect(errors).toEqual([]);
  });

  test('TC-AUTH-06: Valid login redirects to dashboard', async ({ browser }) => {
    const page = await browser.newPage();
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(e.message));
    const responses: string[] = [];
    page.on('response', (resp) => {
      if (resp.status() >= 500) responses.push(`${resp.status()} ${resp.url().substring(0, 60)}`);
    });

    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.fill('#username', 'admin');
    await page.fill('#password', PW);
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);

    // Should NOT be on login page
    expect(page.url()).not.toContain('/login');
    expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
    expect(responses, `API 5xx: ${responses.join(', ')}`).toEqual([]);
  });

  test('TC-AUTH-07: Already logged in bypasses login page', async ({ browser }) => {
    const page = await browser.newPage();
    // Login first
    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.fill('#username', 'admin');
    await page.fill('#password', PW);
    await page.click('button[type="submit"]');
    await page.waitForTimeout(2000);

    // Visit /login again
    const resp = await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    // Should be redirected away from /login
    expect(page.url()).not.toContain('/login');
  });
});

test.describe('AUTH — Redirect & Session', () => {
  test('TC-AUTH-08: Protected page redirects unauthenticated to /login?from=', async ({ browser }) => {
    const page = await browser.newPage();
    await page.goto(`${BASE}/academic/subject-assignments`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    expect(page.url()).toContain('/login');
    const url = new URL(page.url());
    const from = url.searchParams.get('from');
    expect(from).toBeTruthy();
    expect(from).toContain('academic');
  });

  test('TC-AUTH-09: Login after ?from= redirects back to original page', async ({ browser }) => {
    const page = await browser.newPage();
    // Go to protected page → redirected to /login?from=...
    await page.goto(`${BASE}/academic/subject-assignments`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    expect(page.url()).toContain('/login');

    // Login
    await page.fill('#username', 'admin');
    await page.fill('#password', PW);
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);

    // Should redirect back to subject-assignments
    expect(page.url()).toContain('/academic/subject-assignments');
  });

  test('TC-AUTH-10: Complete logout clears session', async ({ browser }) => {
    const page = await browser.newPage();

    // Login first
    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.fill('#username', 'admin');
    await page.fill('#password', PW);
    await page.click('button[type="submit"]');
    await page.waitForTimeout(2000);

    // Logout
    await page.goto(`${BASE}/logout`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    // Should be redirected to landing page
    expect(page.url()).toBe(`${BASE}/`);

    // Access protected page → should redirect to login
    await page.goto(`${BASE}/academic/subject-assignments`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    expect(page.url()).toContain('/login');
  });
});

test.describe('AUTH — Mobile', () => {
  test('TC-AUTH-11: Mobile login fits viewport without horizontal scroll', async ({ browser }) => {
    const page = await browser.newPage();
    await page.setViewportSize({ width: 390, height: 844 });

    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth);
    const viewportWidth = await page.evaluate(() => window.innerWidth);
    expect(scrollWidth).toBeLessThanOrEqual(viewportWidth + 5);
  });

  test('TC-AUTH-12: Mobile login has accessible touch targets', async ({ browser }) => {
    const page = await browser.newPage();
    await page.setViewportSize({ width: 390, height: 844 });

    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    // Check submit button meets touch target size
    const submitBtn = page.locator('button[type="submit"]');
    const box = await submitBtn.boundingBox();
    expect(box).not.toBeNull();
    if (box) {
      expect(box.width).toBeGreaterThanOrEqual(44);
      expect(box.height).toBeGreaterThanOrEqual(44);
    }

    // Check input fields are touch-friendly
    const usernameInput = page.locator('#username');
    const uBox = await usernameInput.boundingBox();
    expect(uBox).not.toBeNull();
    if (uBox) {
      expect(uBox.height).toBeGreaterThanOrEqual(40);
    }
  });

  test('TC-AUTH-13: Mobile login form submission works', async ({ browser }) => {
    const page = await browser.newPage();
    await page.setViewportSize({ width: 390, height: 844 });
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(e.message));

    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.fill('#username', 'admin');
    await page.fill('#password', PW);
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);

    expect(page.url()).not.toContain('/login');
    expect(errors).toEqual([]);
  });
});

test.describe('AUTH — Edge Cases', () => {
  test('TC-AUTH-14: No JS errors on any auth page', async ({ browser }) => {
    const page = await browser.newPage();
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(e.message));

    await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    await page.goto(`${BASE}/`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    await page.goto(`${BASE}/logout`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    // Even 404 pages should not break
    await page.goto(`${BASE}/login?foo=bar`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);

    expect(errors).toEqual([]);
  });

  test('TC-AUTH-15: Login with from= pointing to login is rejected (prevent redirect loop)', async ({ browser }) => {
    const page = await browser.newPage();

    // Attempt redirect loop: from=/login
    const resp = await page.goto(`${BASE}/login?from=%2Flogin`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1000);

    // Should still see login page
    await expect(page.locator('#username')).toBeVisible();

    // Login with the loop redirect
    await page.fill('#username', 'admin');
    await page.fill('#password', PW);
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);

    // Should redirect to / not /login (safe fallback)
    expect(page.url()).not.toContain('/login');
  });

  test('TC-AUTH-16: Login with from= pointing to external is rejected (open redirect prevention)', async ({ browser }) => {
    const page = await browser.newPage();

    await page.goto(`${BASE}/login?from=https%3A%2F%2Fevil.com`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1000);

    await page.fill('#username', 'admin');
    await page.fill('#password', PW);
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);

    // Should NOT redirect to evil.com
    expect(page.url()).not.toContain('evil.com');
    // Should be on the app
    expect(page.url()).toContain('localhost:8021');
  });
});
