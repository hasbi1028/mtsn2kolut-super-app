import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/parents/${id}/children`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'parents/[id]/children GET');
	}
};
