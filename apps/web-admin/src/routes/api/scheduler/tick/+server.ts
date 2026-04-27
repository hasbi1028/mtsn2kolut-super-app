import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPost, handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async () => {
	try {
		const result = await apiPost<{ checked_at: string; processed: number; enqueued: number; skipped: number }>(
			'/api/scheduler/tick'
		);
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'scheduler/tick');
	}
};
