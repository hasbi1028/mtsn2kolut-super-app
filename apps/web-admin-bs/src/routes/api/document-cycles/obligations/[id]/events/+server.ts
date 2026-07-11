import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, apiPathWithQuery, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPathWithQuery(apiPath`/api/document-cycles/obligations/${id}/events`, event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'document-cycles/obligations/[id]/events GET');
	}
};
