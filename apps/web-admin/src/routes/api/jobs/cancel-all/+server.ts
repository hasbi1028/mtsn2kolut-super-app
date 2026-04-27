import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPost, handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async () => {
	try {
		const result = await apiPost<{ ok: boolean; cancelled: number }>('/api/jobs/cancel-all');
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'jobs/cancel-all');
	}
};
