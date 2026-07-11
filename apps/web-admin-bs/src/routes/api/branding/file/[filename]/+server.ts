import type { RequestHandler } from './$types';
import { apiPath, handleRouteError, requiredRouteParam, streamProxyResponse } from '$lib/server/api';
import { env } from '$env/dynamic/private';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const GET: RequestHandler = async ({ fetch, params }) => {
	try {
		const filename = requiredRouteParam(params.filename, 'filename');
		const response = await fetch(`${BASE}${apiPath`/api/branding/file/${filename}`}`);
		return await streamProxyResponse(response, {
			defaultContentType: 'image/png',
			defaultCacheControl: 'public, max-age=31536000, immutable'
		});
	} catch (e) {
		return handleRouteError(e, 'GET /api/branding/file/[filename]');
	}
};
