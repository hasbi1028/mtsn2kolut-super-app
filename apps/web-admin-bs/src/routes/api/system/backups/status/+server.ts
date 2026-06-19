import type { RequestHandler } from './$types';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get('/api/system/backups/status');
		return Response.json({ data });
	} catch (error) {
		return handleRouteError(error, 'system/backups/status GET');
	}
};
