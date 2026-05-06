import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const pid = requiredRouteParam(event.params.pid, 'pid');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(cbtApiPath`/sessions/${id}/participants/${pid}/seat`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/participants/[pid]/seat POST');
	}
};
