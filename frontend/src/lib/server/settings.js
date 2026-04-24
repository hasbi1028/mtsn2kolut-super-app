import { db } from '$lib/server/db';

export function getAppSettings() {
  const rows = db.prepare('SELECT key, value FROM app_settings').all();
  const map = Object.fromEntries(rows.map((row) => [row.key, row.value]));

  return {
    max_concurrent: Math.max(1, Number(map.max_concurrent || 1)),
    headless: map.headless === '1' || map.headless === 'true'
  };
}

export function updateAppSettings(input = {}) {
  const next = getAppSettings();

  if (input.max_concurrent !== undefined) {
    next.max_concurrent = Math.max(1, Math.min(20, Number(input.max_concurrent) || 1));
  }
  if (input.headless !== undefined) {
    next.headless = Boolean(input.headless);
  }

  const tx = db.transaction(() => {
    db.prepare(
      `INSERT INTO app_settings (key, value, updated_at)
       VALUES (?, ?, CURRENT_TIMESTAMP)
       ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP`
    ).run('max_concurrent', String(next.max_concurrent));
    db.prepare(
      `INSERT INTO app_settings (key, value, updated_at)
       VALUES (?, ?, CURRENT_TIMESTAMP)
       ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP`
    ).run('headless', next.headless ? '1' : '0');
  });

  tx();
  return next;
}
