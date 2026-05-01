import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PATCH = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).patch(apiPath`/api/cbt/questions/${id}/workflow`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/[id]/workflow PATCH');
	}
};
