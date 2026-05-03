import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy, streamProxyResponse } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const res = await proxy(event).fetch(apiPathWithQuery('/api/cbt/questions/export', event.url.searchParams));
		return await streamProxyResponse(res, {
			defaultContentType: 'text/csv; charset=utf-8',
			defaultCacheControl: 'no-store',
		});
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/export GET');
	}
};
