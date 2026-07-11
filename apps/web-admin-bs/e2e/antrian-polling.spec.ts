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

test('antrian page auto-refresh polling works', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));

  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.fill('#username', 'admin');
  await page.fill('#password', getAdminPassword());
  await page.click('button[type="submit"]');
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 15000 });

  await page.goto(`${BASE}/pusaka/antrian`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2000);

  // 1. Check "Auto: 30s" badge appears
  const autoBadge = page.locator('text=/Auto:\\s*30s/');
  await expect(autoBadge).toBeVisible();
  console.log('✓ Auto-refresh badge visible');

  // 2. Capture initial "last refresh" timestamp
  const timeSpans = page.locator('span[title="Update terakhir"]');
  await expect(timeSpans).toBeVisible();
  const initialTime = await timeSpans.first().innerText();
  console.log(`Initial time: ${initialTime}`);

  // 3. Wait 32 seconds (slightly more than polling interval) WITHOUT clicking refresh
  console.log('Waiting 32 seconds for auto-refresh...');
  await page.waitForTimeout(32000);

  // 4. Capture new timestamp - it should have changed
  const newTime = await timeSpans.first().innerText();
  console.log(`New time: ${newTime}`);

  // 5. Timestamp must be different (proving polling worked)
  expect(newTime).not.toBe(initialTime);
  console.log('✓ Timestamp changed - polling is working!');

  // 6. Stats should still be visible (not broken by polling)
  const suksesBadge = page.locator('span:has-text("Sukses:")');
  await expect(suksesBadge).toBeVisible();
  console.log('✓ Stats still visible after polling');

  // 7. No JS errors during polling
  expect(errors, `JS errors: ${errors.join(', ')}`).toEqual([]);

  // 8. Take final screenshot
  await page.screenshot({ path: '/tmp/antrian-polling-verify.png', fullPage: true });
  console.log('Screenshot: /tmp/antrian-polling-verify.png');
}, 90_000); // timeout 90s untuk akomodasi wait 32s
