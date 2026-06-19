import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get('/api/notifications/unread-count');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'notifications/unread-count GET');
	}
};
