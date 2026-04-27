import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { handleRouteError } from '$lib/server/api';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const INTERNAL_KEY = env.INTERNAL_API_KEY ?? '';

interface WorkerEntry {
	worker_id: string;
	active_consumers: number;
	target_concurrency: number;
	headless: boolean;
	last_sync_at: string;
	reported_at: string;
}

export const GET: RequestHandler = async () => {
	try {
		const res = await fetch(`${BASE}/health`, {
			headers: { 'X-Internal-Key': INTERNAL_KEY },
		});
		if (!res.ok) throw new Error(`health check HTTP ${res.status}`);
		const data = await res.json();

		const workersRaw: Record<string, string> = data.workers ?? {};
		const active_workers: WorkerEntry[] = Object.values(workersRaw)
			.map((v) => { try { return JSON.parse(v); } catch { return null; } })
			.filter(Boolean) as WorkerEntry[];

		return json({
			data: {
				active_workers,
				total: active_workers.length,
				queue: data.queue ?? {},
				last_checked: new Date().toISOString(),
			},
		});
	} catch (e) {
		return handleRouteError(e, 'worker/status');
	}
};
