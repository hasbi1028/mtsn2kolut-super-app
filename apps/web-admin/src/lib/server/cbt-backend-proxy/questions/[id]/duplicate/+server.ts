import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).post(cbtApiPath`/questions/${id}/duplicate`);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/[id]/duplicate POST');
	}
};
