import { apiGet, apiPost } from './api.js';
import { logError, logInfo } from './logger.js';

const TZ = 'Asia/Makassar';

function nowMakassar(): { date: string; time: string } {
	const d = new Date();
	const date = new Intl.DateTimeFormat('en-CA', { timeZone: TZ, year: 'numeric', month: '2-digit', day: '2-digit' }).format(d);
	const time = new Intl.DateTimeFormat('en-GB', { timeZone: TZ, hour: '2-digit', minute: '2-digit', hour12: false }).format(d);
	return { date, time };
}

declare global {
	// eslint-disable-next-line no-var
	var __scheduleRunCache:     Set<string> | undefined;
	// eslint-disable-next-line no-var
	var __schedulerInitialized: boolean | undefined;
}

interface GoSchedule {
	id: string;
	run_type: string;
	run_time: string;
	is_enabled: boolean;
}

export async function schedulerTick(): Promise<number> {
	const { date, time } = nowMakassar();

	let schedules: GoSchedule[];
	try {
		schedules = await apiGet<GoSchedule[]>('/api/schedules');
	} catch (e) {
		logError('scheduler: failed to fetch schedules', { error: (e as Error)?.message ?? String(e) });
		return 0;
	}

	globalThis.__scheduleRunCache ??= new Set();

	let totalEnqueued = 0;
	for (const s of schedules) {
		if (!s.is_enabled || s.run_time !== time) continue;
		const key = `${date}:${s.id}:${s.run_time}`;
		if (globalThis.__scheduleRunCache.has(key)) continue;

		try {
			const result = await apiPost<{ inserted: number; skipped: number }>(
				'/api/jobs/run-all',
				{ run_type: s.run_type, max_attempts: 3 }
			);
			totalEnqueued += result?.inserted ?? 0;
			globalThis.__scheduleRunCache.add(key);
			logInfo('scheduler enqueued jobs', { schedule_id: s.id, run_type: s.run_type, ...result, at: `${date} ${time}` });
		} catch (e) {
			logError('scheduler: failed to enqueue jobs', { schedule_id: s.id, error: (e as Error)?.message ?? String(e) });
		}
	}

	return totalEnqueued;
}

export function initScheduler() {
	if (globalThis.__schedulerInitialized) return;
	globalThis.__schedulerInitialized = true;

	setInterval(() => {
		schedulerTick().catch((e) => {
			logError('scheduler tick failed', { error: (e as Error)?.message ?? String(e) });
		});
	}, 30_000);

	logInfo('scheduler initialized', {});
}
