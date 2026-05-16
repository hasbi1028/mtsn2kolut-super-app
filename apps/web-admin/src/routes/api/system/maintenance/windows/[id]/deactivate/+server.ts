import type { RequestHandler } from './$types';
import { apiPath, handleRouteError, proxy, readOptionalRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readOptionalRequestJson<Record<string, unknown>>(event.request, {});
		const data = await proxy(event).post(apiPath`/api/system/maintenance/windows/${id}/deactivate`, body);
		return Response.json({ data });
	} catch (error) {
		return handleRouteError(error, 'system/maintenance/windows/[id]/deactivate POST');
	}
};
