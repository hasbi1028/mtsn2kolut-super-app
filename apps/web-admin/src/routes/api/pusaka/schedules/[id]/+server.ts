import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const sched = await proxy(event).put(apiPath`/api/pusaka/schedules/${id}`, body);
		return json(sched);
	} catch (e) {
		return handleRouteError(e, 'pusaka/schedules/[id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const result = await proxy(event).del(apiPath`/api/pusaka/schedules/${id}`);
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'pusaka/schedules/[id] DELETE');
	}
};
