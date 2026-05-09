import type { RequestHandler } from './$types';
import { apiPathWithQuery, handleRouteError, proxy, streamProxyResponse } from '$lib/server/api';

const ALLOWED_EXPORT_QUERY_KEYS = ['event_group', 'event_name', 'source_surface', 'role', 'result', 'days', 'limit', 'offset'] as const;

export const GET: RequestHandler = async (event) => {
	try {
		const params = new URLSearchParams();
		for (const key of ALLOWED_EXPORT_QUERY_KEYS) {
			const value = event.url.searchParams.get(key);
			if (value) params.set(key, value);
		}
		const response = await proxy(event).fetch(apiPathWithQuery('/api/internal-analytics/export', params));
		return await streamProxyResponse(response, {
			defaultContentType: 'text/csv; charset=utf-8',
			defaultCacheControl: 'no-store'
		});
	} catch (e) {
		return handleRouteError(e, 'internal-analytics/export GET');
	}
};
