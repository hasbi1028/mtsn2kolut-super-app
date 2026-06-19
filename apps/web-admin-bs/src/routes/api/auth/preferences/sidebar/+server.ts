import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { proxy, handleRouteError, readRequestJson } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}
	try {
		const data = await proxy(event).get('/api/auth/preferences/sidebar');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'auth/preferences/sidebar GET');
	}
};

export const PATCH: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).patch('/api/auth/preferences/sidebar', body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'auth/preferences/sidebar PATCH');
	}
};
