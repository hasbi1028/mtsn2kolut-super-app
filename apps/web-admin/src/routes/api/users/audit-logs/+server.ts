import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const qs = new URLSearchParams();
		const page = event.url.searchParams.get('page');
		const perPage = event.url.searchParams.get('per_page');
		if (page) qs.set('page', page);
		if (perPage) qs.set('per_page', perPage);

		const res = await proxy(event).get(`/api/users/audit-logs?${qs}`);
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'audit-logs GET');
	}
};
