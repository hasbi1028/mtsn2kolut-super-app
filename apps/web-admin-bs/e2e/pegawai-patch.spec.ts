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

test('pegawai pusaka page shows data after patch', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));

  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', getAdminPassword());
  await page.click('button[type="submit"]');
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 15000 });

  await page.goto(`${BASE}/pusaka/employees`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Check headers
  const headers = await page.locator('table thead th').allTextContents();
  console.log(`Headers: ${JSON.stringify(headers)}`);
  expect(headers).toEqual(['#', 'Nama', 'NIP', 'Unit Kerja', 'Status', 'Pusaka']);

  // Check: at least 1 row
  const rowCount = await page.locator('table tbody tr').count();
  console.log(`Row count: ${rowCount}`);
  expect(rowCount).toBeGreaterThan(0);

  // Check: counter should be > 0
  const counterText = await page.locator('span.fw-bold.small').first().innerText();
  console.log(`Counter: ${counterText}`);
  expect(counterText).not.toContain('0 pegawai');

  // Check: first row Nama should not be "—"
  const firstNama = await page.locator('table tbody tr').first().locator('td').nth(1).innerText();
  console.log(`First row Nama: ${firstNama}`);
  expect(firstNama).not.toBe('—');
  expect(firstNama.length).toBeGreaterThan(3);

  // Check: first row NIP should not be "—"
  const firstNIP = await page.locator('table tbody tr').first().locator('td').nth(2).innerText();
  console.log(`First row NIP: ${firstNIP}`);
  expect(firstNIP).not.toBe('—');

  // No JS errors
  expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);

  await page.screenshot({ path: '/tmp/pegawai-after-patch.png', fullPage: true });
  console.log('Screenshot: /tmp/pegawai-after-patch.png');
});
