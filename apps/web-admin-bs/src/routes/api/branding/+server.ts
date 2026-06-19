import type { RequestHandler } from './$types';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';
import type { BrandingSettings } from '$lib/branding';

export const GET: RequestHandler = async (event) => {
	try {
		return Response.json(await proxy(event).get<BrandingSettings>('/api/branding'));
	} catch (e) {
		return handleRouteError(e, 'GET /api/branding');
	}
};

export const PUT: RequestHandler = async (event) => {
	try {
		const body = await readRequestJson<BrandingSettings>(event.request, 64 * 1024);
		return Response.json(await proxy(event).put<BrandingSettings>('/api/branding', body));
	} catch (e) {
		return handleRouteError(e, 'PUT /api/branding');
	}
};
