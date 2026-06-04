import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).post(apiPath`/api/cbt/sessions/${id}/finalize-overdue`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/sessions/[id]/finalize-overdue POST');
	}
};
