import type { RequestHandler } from './$types';
import { apiPath, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).post(apiPath`/api/system/backups/${id}/validate-restore`, {});
		return Response.json({ data });
	} catch (error) {
		return handleRouteError(error, 'system/backups/[id]/validate-restore POST');
	}
};
