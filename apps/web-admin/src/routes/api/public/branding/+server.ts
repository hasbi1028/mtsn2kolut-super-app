import type { RequestHandler } from './$types';
import { handleRouteError } from '$lib/server/api';
import { normalizeBranding, defaultBranding } from '$lib/branding';
import { env } from '$env/dynamic/private';

const API_BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export const GET: RequestHandler = async (event) => {
	try {
		const r = await event.fetch(`${API_BASE}/api/public/branding`);
		if (r.ok) {
			const payload = await r.json();
			const data = payload?.data ?? payload ?? {};
			// normalizeBranding expects the full branding object at top level
			return Response.json(normalizeBranding(data));
		}
		return Response.json(defaultBranding);
	} catch (e) {
		return handleRouteError(e, 'GET /api/public/branding');
	}
};
