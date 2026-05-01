import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPathWithQuery, proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/kesiswaan/students', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/students GET');
	}
};
