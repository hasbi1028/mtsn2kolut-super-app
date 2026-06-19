import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/governance/documents', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'governance/documents GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/governance/documents', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'governance/documents POST');
	}
};
