import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

type WorkerInfo = {
	worker_id: string;
	status: 'active' | 'stale' | 'offline';
	active_consumers: number;
	target_concurrency: number;
	headless: boolean;
	last_sync_at: string;
	reported_at: string;
};

type WorkerStatusPayload = {
	workers: WorkerInfo[];
	total: number;
	active_count: number;
	global_max_concurrent: string;
	queue?: {
		queued: number;
		running: number;
		success: number;
		failed: number;
	};
	last_checked: string;
};

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get<WorkerStatusPayload>('/api/pusaka/worker/status');
		return json({
			data: {
				workers: data.workers ?? [],
				total: data.total ?? 0,
				active_count: data.active_count ?? 0,
				global_max_concurrent: data.global_max_concurrent ?? '',
				queue: data.queue ?? { queued: 0, running: 0, success: 0, failed: 0 },
				last_checked: data.last_checked ?? '',
			},
		});
	} catch (e) {
		return handleRouteError(e, 'pusaka/worker/status');
	}
};
