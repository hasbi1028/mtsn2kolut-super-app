import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/library/books/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'library/books PUT');
	}
};

export const DELETE: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(apiPath`/api/library/books/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'library/books DELETE');
	}
};
