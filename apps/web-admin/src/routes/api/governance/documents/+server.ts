import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const params = new URLSearchParams(event.url.searchParams);
		const data = await proxy(event).get(`/api/governance/documents?${params.toString()}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'governance/documents GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post('/api/governance/documents', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'governance/documents POST');
	}
};
