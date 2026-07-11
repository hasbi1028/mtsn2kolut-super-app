import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPathWithQuery, proxy, handleRouteError, readRequestJson } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/library/loans', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'library/loans GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/library/loans', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'library/loans POST');
	}
};
