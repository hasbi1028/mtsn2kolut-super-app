import { eq } from 'drizzle-orm';
import { db } from './db.js';
import { appSettings } from './schema.js';

export interface AppSettings {
	max_concurrent: number;
	headless: boolean;
}

export function getAppSettings(): AppSettings {
	const rows = db.select().from(appSettings).all();
	const map  = Object.fromEntries(rows.map((r) => [r.key, r.value]));
	return {
		max_concurrent: Math.max(1, Number(map.max_concurrent ?? 1)),
		headless: map.headless === '1' || map.headless === 'true',
	};
}

export function updateAppSettings(input: Partial<AppSettings>): AppSettings {
	const next = getAppSettings();

	if (input.max_concurrent !== undefined) {
		next.max_concurrent = Math.max(1, Math.min(20, Number(input.max_concurrent) || 1));
	}
	if (input.headless !== undefined) {
		next.headless = Boolean(input.headless);
	}

	db.transaction((tx) => {
		tx.insert(appSettings)
			.values({ key: 'max_concurrent', value: String(next.max_concurrent) })
			.onConflictDoUpdate({ target: appSettings.key, set: { value: String(next.max_concurrent), updated_at: new Date().toISOString() } })
			.run();
		tx.insert(appSettings)
			.values({ key: 'headless', value: next.headless ? '1' : '0' })
			.onConflictDoUpdate({ target: appSettings.key, set: { value: next.headless ? '1' : '0', updated_at: new Date().toISOString() } })
			.run();
	});

	return next;
}
