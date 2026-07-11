import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, jsonProxyResponse, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const res = await proxy(event).fetch(apiPathWithQuery('/api/pusaka/attendance', event.url.searchParams));
		return await jsonProxyResponse<Record<string, unknown>>(res);
	} catch (e) {
		return handleRouteError(e, 'pusaka/attendance GET');
	}
};
