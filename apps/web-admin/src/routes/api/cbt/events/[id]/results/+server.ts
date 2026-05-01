import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/cbt/events/${id}/results`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/results GET');
	}
};
