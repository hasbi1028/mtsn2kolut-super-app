import { eq } from 'drizzle-orm';
import { db } from './db.js';
import { schedules } from './schema.js';
import { enqueueAllActive } from './queue.js';
import { logError, logInfo } from './logger.js';
import type { RunType } from './schema.js';

const TZ = 'Asia/Makassar';

function nowMakassar(): { date: string; time: string } {
	const d = new Date();
	const date = new Intl.DateTimeFormat('en-CA', { timeZone: TZ, year: 'numeric', month: '2-digit', day: '2-digit' }).format(d);
	const time = new Intl.DateTimeFormat('en-GB', { timeZone: TZ, hour: '2-digit', minute: '2-digit', hour12: false }).format(d);
	return { date, time };
}

declare global {
	// eslint-disable-next-line no-var
	var __scheduleRunCache:    Set<string> | undefined;
	// eslint-disable-next-line no-var
	var __schedulerInitialized: boolean | undefined;
}

export function schedulerTick(): number {
	const { date, time } = nowMakassar();
	const rows = db.select().from(schedules).where(eq(schedules.is_enabled, 1)).all();

	globalThis.__scheduleRunCache ??= new Set();

	let totalEnqueued = 0;
	for (const s of rows) {
		if (s.run_time !== time) continue;
		const key = `${date}:${s.id}:${s.run_time}`;
		if (globalThis.__scheduleRunCache.has(key)) continue;

		const result = enqueueAllActive(s.run_type as RunType, 3);
		totalEnqueued += result.inserted;
		globalThis.__scheduleRunCache.add(key);
		logInfo('scheduler enqueued jobs', { schedule_id: s.id, run_type: s.run_type, ...result, at: `${date} ${time}` });
	}

	return totalEnqueued;
}

export function initScheduler() {
	if (globalThis.__schedulerInitialized) return;
	globalThis.__schedulerInitialized = true;

	setInterval(() => {
		try { schedulerTick(); }
		catch (e) { logError('scheduler tick failed', { error: (e as Error)?.message ?? String(e) }); }
	}, 30_000);

	logInfo('scheduler initialized', {});
}
