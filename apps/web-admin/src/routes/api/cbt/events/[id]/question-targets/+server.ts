import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/cbt/events/${id}/question-targets`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/question-targets GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/cbt/events/${id}/question-targets`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/question-targets PUT');
	}
};
