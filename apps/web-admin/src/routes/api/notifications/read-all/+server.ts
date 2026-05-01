import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).post('/api/notifications/read-all');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'notifications/read-all POST');
	}
};
