import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

type WorkerStatusPayload = {
	active_workers: Record<string, unknown>[];
	total: number;
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
				active_workers: data.active_workers,
				total: data.total,
				queue: data.queue ?? { queued: 0, running: 0, success: 0, failed: 0 },
				last_checked: data.last_checked,
			},
		});
	} catch (e) {
		return handleRouteError(e, 'pusaka/worker/status');
	}
};
