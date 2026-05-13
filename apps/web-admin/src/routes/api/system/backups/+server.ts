import type { RequestHandler } from './$types';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get('/api/system/backups');
		return Response.json({ data });
	} catch (error) {
		return handleRouteError(error, 'system/backups GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json().catch(() => ({}));
		const data = await proxy(event).post('/api/system/backups/run', body);
		return Response.json({ data });
	} catch (error) {
		return handleRouteError(error, 'system/backups POST');
	}
};
