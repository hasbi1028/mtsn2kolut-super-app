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

test('antrian page shows employee data', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));

  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', getAdminPassword());
  await page.click('button[type="submit"]');
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 15000 });

  await page.goto(`${BASE}/pusaka/antrian`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // Check table headers
  const headers = await page.locator('table thead th').allTextContents();
  console.log(`Headers: ${JSON.stringify(headers)}`);
  expect(headers).toEqual(['#', 'Pegawai (NIP)', 'Tipe', 'Status', 'Waktu', '']);

  // Check: should have at least 1 row
  const rowCount = await page.locator('table tbody tr').count();
  console.log(`Row count: ${rowCount}`);
  expect(rowCount).toBeGreaterThan(0);

  // Check: at least one cell should contain a real name (not just "—")
  const firstRow = page.locator('table tbody tr').first();
  const cellText = await firstRow.locator('td').nth(1).innerText();
  console.log(`First row Pegawai cell: ${cellText}`);
  expect(cellText).not.toBe('—');
  expect(cellText.length).toBeGreaterThan(3);

  // Check: column "Nama" should NOT exist
  const namaCol = await page.locator('th:has-text("Nama")').count();
  expect(namaCol, 'Kolom Nama (lama) should be removed').toBe(0);

  // Check: stats badges
  const suksesBadge = await page.locator('span:has-text("Sukses:")').count();
  expect(suksesBadge).toBeGreaterThan(0);

  // No JS errors
  expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);

  await page.screenshot({ path: '/tmp/antrian-after-patch.png', fullPage: true });
  console.log('Screenshot saved');
});
