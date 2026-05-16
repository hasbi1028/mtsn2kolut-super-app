import type { RequestHandler } from './$types';
import { apiPath, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const PUT: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await event.request.json();
		const data = await proxy(event).put(apiPath`/api/system/maintenance/windows/${id}`, body);
		return Response.json({ data });
	} catch (error) {
		return handleRouteError(error, 'system/maintenance/windows/[id] PUT');
	}
};
