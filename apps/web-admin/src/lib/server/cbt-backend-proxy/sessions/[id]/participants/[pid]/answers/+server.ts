import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const pid = requiredRouteParam(event.params.pid, 'pid');
		const data = await proxy(event).get(cbtApiPath`/sessions/${id}/participants/${pid}/answers`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/participants/[pid]/answers GET');
	}
};
