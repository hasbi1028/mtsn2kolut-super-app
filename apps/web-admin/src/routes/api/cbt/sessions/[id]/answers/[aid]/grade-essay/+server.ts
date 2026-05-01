import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const aid = requiredRouteParam(event.params.aid, 'aid');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(apiPath`/api/cbt/sessions/${id}/answers/${aid}/grade-essay`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/answers/[aid]/grade-essay POST');
	}
};
