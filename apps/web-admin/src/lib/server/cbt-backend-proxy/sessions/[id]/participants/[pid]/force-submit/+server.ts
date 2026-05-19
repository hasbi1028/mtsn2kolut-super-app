import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const pid = requiredRouteParam(event.params.pid, 'pid');
		const body = await readRequestJson<{ notes?: string }>(event.request, 8 * 1024);
		const data = await proxy(event).post(cbtApiPath`/sessions/${id}/participants/${pid}/force-submit`, {
			notes: typeof body.notes === 'string' ? body.notes.trim() : '',
		});
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/participants/[pid]/force-submit POST');
	}
};
