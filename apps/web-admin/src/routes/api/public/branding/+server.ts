import type { RequestHandler } from './$types';
import { apiPublicGetWithFetch, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async ({ fetch }) => {
	try {
		return Response.json(await apiPublicGetWithFetch(fetch, '/api/public/branding'));
	} catch (e) {
		return handleRouteError(e, 'GET /api/public/branding');
	}
};
