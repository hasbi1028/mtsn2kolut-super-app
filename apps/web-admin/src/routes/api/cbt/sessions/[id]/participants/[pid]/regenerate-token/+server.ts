import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const pid = requiredRouteParam(event.params.pid, 'pid');
		const data = await proxy(event).post(apiPath`/api/cbt/sessions/${id}/participants/${pid}/regenerate-token`, {});
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/participants/[pid]/regenerate-token POST');
	}
};
