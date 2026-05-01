import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/governance/work-plan-items', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'governance/work-plan-items GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/governance/work-plan-items', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'governance/work-plan-items POST');
	}
};
