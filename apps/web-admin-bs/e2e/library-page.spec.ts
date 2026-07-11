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

test('library page works', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));

  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', getAdminPassword());
  await page.click('button[type="submit"]');
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 15000 });

  // Test library main page
  await page.goto(`${BASE}/library`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);
  const title = await page.title();
  console.log('Title:', title);
  expect(title).toContain('Perpustakaan');

  // Test "Tambah Buku" form opens
  const tambahBtn = page.locator('button:has-text("Tambah Buku")');
  await expect(tambahBtn).toBeVisible();
  await tambahBtn.click();
  await page.waitForTimeout(500);

  // Form fields
  const kodeInput = page.locator('input[placeholder="BK-001"]');
  await expect(kodeInput).toBeVisible();
  console.log('Form opened OK');

  // Close form
  await page.click('button[aria-label="Tutup"]');
  await page.waitForTimeout(300);

  // Test navigation to loans page
  await page.click('a:has-text("Peminjaman")');
  await page.waitForURL((u) => u.pathname.includes('/loans'), { timeout: 5000 });
  await page.waitForTimeout(1500);
  const loansTitle = await page.title();
  console.log('Loans title:', loansTitle);
  expect(loansTitle).toContain('Peminjaman');

  await page.screenshot({ path: '/tmp/library-page.png', fullPage: true });
  console.log('Screenshot: /tmp/library-page.png');

  expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
});
