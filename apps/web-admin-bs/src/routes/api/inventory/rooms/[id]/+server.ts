import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/inventory/rooms/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'inventory/rooms/[id] GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/inventory/rooms/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'inventory/rooms/[id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(apiPath`/api/inventory/rooms/${id}`);
		return json({ ok: true });
	} catch (e) {
		return handleRouteError(e, 'inventory/rooms/[id] DELETE');
	}
};
