import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const query = event.url.searchParams.toString();
		const data = await proxy(event).get(`/api/academic/curriculum${query ? `?${query}` : ''}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic curriculum GET');
	}
};
