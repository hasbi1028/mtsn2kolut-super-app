import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { proxy, apiPathWithQuery, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/users/change-requests', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'users/change-requests GET');
	}
};
