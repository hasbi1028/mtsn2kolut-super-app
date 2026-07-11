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

test('telegram-laporan page after patch', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));

  // Login
  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', getAdminPassword());
  await page.click('button[type="submit"]');
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 15000 });

  // Navigate to telegram-laporan
  await page.goto(`${BASE}/pusaka/telegram-laporan`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1500);

  // Check: textarea "Pesan" should NOT exist
  const textareaCount = await page.locator('textarea').count();
  console.log(`Textarea count: ${textareaCount} (expected 0)`);
  expect(textareaCount, 'Pesan textarea should be removed').toBe(0);

  // Check: label "Pesan (opsional)" should NOT exist
  const pesanLabel = await page.locator('label:has-text("Pesan (opsional)")').count();
  console.log(`Label "Pesan (opsional)": ${pesanLabel} (expected 0)`);
  expect(pesanLabel, 'Pesan label should be removed').toBe(0);

  // Check: Kirim Laporan button should still exist
  const buttonVisible = await page.locator('button:has-text("Kirim Laporan")').isVisible();
  expect(buttonVisible, 'Kirim Laporan button should still exist').toBe(true);

  // Check: table headers should be # | Tanggal | Waktu | Status (4 columns)
  const headers = await page.locator('table thead th').allTextContents();
  console.log(`Table headers: ${JSON.stringify(headers)}`);
  expect(headers).toEqual(['#', 'Tanggal', 'Waktu', 'Status']);

  // No JS errors
  expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);

  // Screenshot for visual verification
  await page.screenshot({ path: '/tmp/telegram-laporan-after-patch.png', fullPage: true });
  console.log('Screenshot saved: /tmp/telegram-laporan-after-patch.png');
});
