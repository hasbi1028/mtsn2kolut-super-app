import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, apiPathWithQuery, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const qs = new URLSearchParams();
		const page = event.url.searchParams.get('page');
		const perPage = event.url.searchParams.get('per_page');
		if (page) qs.set('page', page);
		if (perPage) qs.set('per_page', perPage);
		const result = await proxy(event).get(apiPathWithQuery(apiPath`/api/pusaka/employees/${id}/audit-logs`, qs));
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'pusaka/employees/:id/audit-logs GET');
	}
};
