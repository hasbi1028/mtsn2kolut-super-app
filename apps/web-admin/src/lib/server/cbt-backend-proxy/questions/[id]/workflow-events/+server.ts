import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(cbtApiPath`/questions/${id}/workflow-events`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/[id]/workflow-events GET');
	}
};
