import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get('/api/tu/archives/stats');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/archives/stats GET');
	}
};
