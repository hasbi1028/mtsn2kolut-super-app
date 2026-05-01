import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const PATCH: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).patch(`/api/document-cycles/obligations/${event.params.id}/status`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'document-cycles/obligations/[id]/status PATCH');
	}
};
