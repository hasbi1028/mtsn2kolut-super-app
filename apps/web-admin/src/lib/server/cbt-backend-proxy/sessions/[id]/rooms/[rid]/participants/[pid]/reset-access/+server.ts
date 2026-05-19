import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const rid = requiredRouteParam(event.params.rid, 'rid');
		const pid = requiredRouteParam(event.params.pid, 'pid');
		const body = await readRequestJson<{ reason?: string; notes?: string }>(event.request, 8 * 1024);
		const data = await proxy(event).post(cbtApiPath`/sessions/${id}/rooms/${rid}/participants/${pid}/reset-access`, {
			reason: typeof body.reason === 'string' ? body.reason.trim() : '',
			notes: typeof body.notes === 'string' ? body.notes.trim() : '',
		});
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms/[rid]/participants/[pid]/reset-access POST');
	}
};
