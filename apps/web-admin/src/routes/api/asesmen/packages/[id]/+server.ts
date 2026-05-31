import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	const id = requiredRouteParam(event.params.id, 'id');
	try {
		const data = await proxy(event).get(apiPath`/api/asesmen/packages/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/packages/[id] GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	const id = requiredRouteParam(event.params.id, 'id');
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/asesmen/packages/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/packages/[id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	const id = requiredRouteParam(event.params.id, 'id');
	try {
		await proxy(event).del(apiPath`/api/asesmen/packages/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'asesmen/packages/[id] DELETE');
	}
};
