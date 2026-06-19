import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { apiPathWithQuery, handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/users/change-requests/pending-count', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'users/change-requests/pending-count GET');
	}
};
