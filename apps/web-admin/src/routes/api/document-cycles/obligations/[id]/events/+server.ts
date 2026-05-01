import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(`/api/document-cycles/obligations/${event.params.id}/events`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'document-cycles/obligations/[id]/events GET');
	}
};
