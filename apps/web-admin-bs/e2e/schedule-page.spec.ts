import { test, expect } from '@playwright/test';
import * as fs from 'fs';

const BASE = process.env.E2E_BASE_URL || 'http://localhost:8222';

function getAdminPassword(): string {
  const envPath = '/home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api/.env';
  const content = fs.readFileSync(envPath, 'utf-8');
  for (const line of content.split('\n')) {
    if (line.startsWith('ADMIN_PASSWORD=')) {
      return line.substring('ADMIN_PASSWORD='.length).trim();
    }
  }
  throw new Error('ADMIN_PASSWORD not found');
}

test('employee schedule page works', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));

  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', getAdminPassword());
  await page.click('button[type="submit"]');
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 15000 });

  // Go to employees list
  await page.goto(`${BASE}/pusaka/employees`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Check ada tombol atur jadwal di tiap baris
  const scheduleButtons = page.locator('a[aria-label="Atur jadwal"]');
  const count = await scheduleButtons.count();
  console.log(`Schedule buttons: ${count}`);
  expect(count).toBeGreaterThan(0);

  // Get first employee ID dari href tombol
  const firstHref = await scheduleButtons.first().getAttribute('href');
  console.log(`First href: ${firstHref}`);
  expect(firstHref).toMatch(/\/pusaka\/employees\/[a-f0-9-]+\/schedules/);

  // Navigate to schedule page
  await page.goto(`${BASE}${firstHref}`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Check title
  const title = await page.title();
  console.log(`Title: ${title}`);
  expect(title).toContain('Atur Jadwal');

  // Check table ada 7 hari (Senin-Minggu)
  const dayHeaders = await page.locator('table tbody tr td.fw-semibold').allTextContents();
  console.log(`Days: ${dayHeaders.join(', ')}`);
  expect(dayHeaders.length).toBe(7);

  // No JS errors
  expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);

  await page.screenshot({ path: '/tmp/schedule-page.png', fullPage: true });
  console.log('Screenshot: /tmp/schedule-page.png');
});
