import type { RequestHandler } from './$types';
import { apiPublicGetWithFetch, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await apiPublicGetWithFetch(event.fetch, '/api/system/maintenance/status');
		return Response.json({ data });
	} catch (error) {
		return handleRouteError(error, 'system/maintenance/status GET');
	}
};
