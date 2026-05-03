import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const rid = requiredRouteParam(event.params.rid, 'rid');
		const data = await proxy(event).get(apiPath`/api/cbt/sessions/${id}/rooms/${rid}/proctors`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms/[rid]/proctors GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const rid = requiredRouteParam(event.params.rid, 'rid');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/cbt/sessions/${id}/rooms/${rid}/proctors`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms/[rid]/proctors PUT');
	}
};
