import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const params = new URLSearchParams(event.url.searchParams);
		const data = await proxy(event).get(`/api/kesiswaan/violation-categories?${params.toString()}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/violation-categories GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post('/api/kesiswaan/violation-categories', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/violation-categories POST');
	}
};
