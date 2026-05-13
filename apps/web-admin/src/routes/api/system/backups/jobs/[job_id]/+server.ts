import type { RequestHandler } from './$types';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(`/api/system/backups/jobs/${encodeURIComponent(event.params.job_id)}`);
		return Response.json({ data });
	} catch (error) {
		return handleRouteError(error, 'system/backups jobs GET');
	}
};
