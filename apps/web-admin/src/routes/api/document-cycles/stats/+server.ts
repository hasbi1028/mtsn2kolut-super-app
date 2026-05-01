import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const params = new URLSearchParams(event.url.searchParams);
		const data = await proxy(event).get(`/api/document-cycles/stats?${params.toString()}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'document-cycles/stats GET');
	}
};
