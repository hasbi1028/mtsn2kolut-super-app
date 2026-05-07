import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'student_id');
		const data = await proxy(event).get(apiPath`/api/portal/parent/children/${id}/results`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'portal/orang-tua/children/:id/results GET');
	}
};
