import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const params = new URLSearchParams(event.url.searchParams);
		const suffix = params.toString() ? `?${params.toString()}` : '';
		const data = await proxy(event).get(`/api/governance/work-plan-items${suffix}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'governance/work-plan-items GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post('/api/governance/work-plan-items', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'governance/work-plan-items POST');
	}
};
