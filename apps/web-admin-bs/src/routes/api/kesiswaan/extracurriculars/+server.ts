import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/kesiswaan/extracurriculars', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/extracurriculars GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/kesiswaan/extracurriculars', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/extracurriculars POST');
	}
};
