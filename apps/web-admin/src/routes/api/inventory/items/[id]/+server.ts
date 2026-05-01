import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/inventory/items/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'inventory/items PUT');
	}
};

export const DELETE: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(apiPath`/api/inventory/items/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'inventory/items DELETE');
	}
};
