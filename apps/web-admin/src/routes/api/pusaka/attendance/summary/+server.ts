import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, jsonProxyResponse, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const startDate = event.url.searchParams.get('start_date') ?? '';
		const endDate = event.url.searchParams.get('end_date') ?? '';
		const params = new URLSearchParams();
		if (startDate) params.set('start_date', startDate);
		if (endDate) params.set('end_date', endDate);
		const res = await proxy(event).fetch(apiPathWithQuery('/api/pusaka/attendance/summary', params));
		return await jsonProxyResponse<Record<string, unknown>>(res);
	} catch (e) {
		return handleRouteError(e, 'pusaka/attendance/summary GET');
	}
};
