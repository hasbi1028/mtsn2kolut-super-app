import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPost, handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async () => {
	try {
		await apiPost('/api/jobs/cancel-all');
		return json({ ok: true });
	} catch (e) {
		return handleRouteError(e, 'jobs/cancel-all');
	}
};
