import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).post(cbtApiPath`/sessions/${id}/finalize-overdue`, {});
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/sessions finalize-overdue POST');
	}
};
