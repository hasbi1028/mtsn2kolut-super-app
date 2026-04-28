import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const result = await proxy(event).post<{ checked_at: string; processed: number; enqueued: number; skipped: number }>(
			'/api/scheduler/tick'
		);
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'scheduler/tick');
	}
};
