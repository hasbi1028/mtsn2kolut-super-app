import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { ApiError, apiPathWithQuery, handleRouteError, proxy } from '$lib/server/api';

const ALLOWED_SUMMARY_QUERY_KEYS = ['event_group', 'source_surface', 'role', 'result', 'days'] as const;

export const GET: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const params = new URLSearchParams();
		for (const key of ALLOWED_SUMMARY_QUERY_KEYS) {
			const value = event.url.searchParams.get(key);
			if (value) params.set(key, value);
		}
		const data = await proxy(event).get(apiPathWithQuery('/api/internal-analytics/summary', params));
		return json(data);
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'internal-analytics/summary GET');
	}
};
