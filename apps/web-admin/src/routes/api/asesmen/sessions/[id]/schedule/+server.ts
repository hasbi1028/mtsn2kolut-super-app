import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PATCH = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request, 8 << 10);
		const data = await proxy(event).patch(apiPath`/api/cbt/sessions/${id}/schedule`, {
			scheduled_start: body.scheduled_start,
			scheduled_end: body.scheduled_end
		});
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/sessions/[id]/schedule PATCH');
	}
};
