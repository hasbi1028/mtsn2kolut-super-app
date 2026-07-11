import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/document-cycles/obligations/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'document-cycles/obligations/[id] PUT');
	}
};

export const DELETE: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(apiPath`/api/document-cycles/obligations/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'document-cycles/obligations/[id] DELETE');
	}
};
