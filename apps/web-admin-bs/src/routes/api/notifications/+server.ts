import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/notifications', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'notifications GET');
	}
};
