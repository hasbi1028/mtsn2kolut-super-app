import type { RequestHandler } from './$types';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get('/api/system/maintenance/health-summary');
		return Response.json({ data });
	} catch (error) {
		return handleRouteError(error, 'system/maintenance/health-summary GET');
	}
};
