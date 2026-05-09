import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { ApiError, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

const INTERNAL_ANALYTICS_BFF_BODY_LIMIT = 16 * 1024;

export const POST: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request, INTERNAL_ANALYTICS_BFF_BODY_LIMIT);
		const data = await proxy(event).post('/api/internal-analytics/events', body);
		return json(data, { status: 201 });
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'internal-analytics/events POST');
	}
};
