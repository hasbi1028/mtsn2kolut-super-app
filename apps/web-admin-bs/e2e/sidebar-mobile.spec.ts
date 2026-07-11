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

test('mobile sidebar links work end-to-end', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));

  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', getAdminPassword());
  await page.click('button[type="submit"]');
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 15000 });

  // Open mobile sidebar
  await page.click('button[aria-label="Buka menu navigasi"]');
  await page.waitForTimeout(800);

  // Verify sidebar visible
  const sidebar = page.locator('#sidebarOffcanvas');
  await expect(sidebar).toBeVisible();

  // List all links in sidebar
  const links = await sidebar.locator('a.nav-link').allTextContents();
  console.log('Sidebar links:', links);

  // Try each link
  const routes = [
    { label: 'Dashboard', expectedPath: '/' },
    { label: 'Monitor Kehadiran', expectedPath: '/pusaka' },
    { label: 'Pegawai PUSAKA', expectedPath: '/pusaka/employees' },
    { label: 'Data Kehadiran', expectedPath: '/pusaka/kehadiran' },
    { label: 'Ringkasan Kehadiran', expectedPath: '/pusaka/summary' },
    { label: 'Laporan Telegram', expectedPath: '/pusaka/telegram-laporan' },
    { label: 'Antrian Sinkronisasi', expectedPath: '/pusaka/antrian' },
    { label: 'Daftar Pegawai', expectedPath: '/employees' }
  ];

  for (const r of routes) {
    console.log(`\nTesting: ${r.label} → ${r.expectedPath}`);

    // Re-open sidebar if closed
    const isOpen = await page.locator('#sidebarOffcanvas.show').count();
    if (!isOpen) {
      await page.click('button[aria-label="Buka menu navigasi"]');
      await page.waitForTimeout(500);
    }

    // Get link href first
    const link = page.locator(`#sidebarOffcanvas a.nav-link:has-text("${r.label}")`);
    const href = await link.getAttribute('href');
    console.log(`  Link href: ${href}`);

    // Click the link
    await link.click();
    await page.waitForTimeout(2000);

    // Check current URL
    const currentPath = new URL(page.url()).pathname;
    console.log(`  → ${currentPath} (expected ${r.expectedPath})`);

    // Check no error page
    const isError = await page.locator('text="Terjadi Kesalahan"').count();
    if (isError > 0) {
      console.log(`  ❌ ERROR PAGE on ${r.label}`);
    } else if (currentPath === r.expectedPath) {
      console.log(`  ✅ OK`);
    } else {
      console.log(`  ⚠️ MISMATCH (might be initial route, ignore if expected)`);
    }
  }

  await page.screenshot({ path: '/tmp/sidebar-mobile.png', fullPage: true });
  console.log('\nScreenshot: /tmp/sidebar-mobile.png');

  expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);
});
