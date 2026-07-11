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

test('mobile sidebar navigation works after fix', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));

  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', getAdminPassword());
  await page.click('button[type="submit"]');
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 15000 });

  // Test 3 critical links
  const tests = [
    { label: 'Pegawai PUSAKA', expectedPath: '/pusaka/employees' },
    { label: 'Antrian Sinkronisasi', expectedPath: '/pusaka/antrian' },
    { label: 'Laporan Telegram', expectedPath: '/pusaka/telegram-laporan' }
  ];

  for (const t of tests) {
    console.log(`\n=== Testing: ${t.label} → ${t.expectedPath} ===`);

    // Open sidebar
    await page.click('button[aria-label="Buka menu navigasi"]');
    await page.waitForTimeout(500);

    // Click link
    const link = page.locator(`#sidebarOffcanvas a.nav-link:has-text("${t.label}")`);
    await link.click();
    await page.waitForTimeout(1500);

    // Check URL
    const currentPath = new URL(page.url()).pathname;
    if (currentPath === t.expectedPath) {
      console.log(`  ✅ Navigated to ${currentPath}`);
    } else {
      console.log(`  ❌ Still at ${currentPath} (expected ${t.expectedPath})`);
    }
    expect(currentPath).toBe(t.expectedPath);
  }

  expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);

  await page.screenshot({ path: '/tmp/sidebar-fixed.png', fullPage: true });
  console.log('\nScreenshot: /tmp/sidebar-fixed.png');
});
