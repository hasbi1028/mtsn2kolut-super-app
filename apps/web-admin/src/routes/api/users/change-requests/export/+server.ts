import type { RequestHandler } from './$types';
import { apiPathWithQuery, handleRouteError, proxy, streamProxyResponse } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const response = await proxy(event).fetch(apiPathWithQuery('/api/users/change-requests/export', event.url.searchParams));
		return await streamProxyResponse(response, {
			defaultContentType: 'text/csv; charset=utf-8',
			defaultCacheControl: 'no-store'
		});
	} catch (e) {
		return handleRouteError(e, 'users/change-requests/export GET');
	}
};
