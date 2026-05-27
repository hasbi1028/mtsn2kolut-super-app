import type { RequestHandler } from './$types';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(`/api/system/maintenance/windows${event.url.search}`);
		return Response.json({ data });
	} catch (error) {
		return handleRouteError(error, 'system/maintenance/windows GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/system/maintenance/windows', body);
		return Response.json({ data }, { status: 201 });
	} catch (error) {
		return handleRouteError(error, 'system/maintenance/windows POST');
	}
};
