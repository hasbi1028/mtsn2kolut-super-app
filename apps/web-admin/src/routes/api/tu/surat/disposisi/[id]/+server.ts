import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/tu/surat/disposisi/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat/disposisi/[id] GET');
	}
};

export const PUT: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/tu/surat/disposisi/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat/disposisi/[id] PUT');
	}
};

export const DELETE: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(apiPath`/api/tu/surat/disposisi/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'tu/surat/disposisi/[id] DELETE');
	}
};
