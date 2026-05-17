import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const pid = requiredRouteParam(event.params.pid, 'pid');
		const body = await readRequestJson(event.request);
		const data = await proxy(event).post(cbtApiPath`/sessions/${id}/participants/${pid}/proctor-actions`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/participants/[pid]/proctor-actions POST');
	}
};
