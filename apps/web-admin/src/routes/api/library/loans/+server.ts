import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const status = event.url.searchParams.get('status') ?? '';
		const data = await proxy(event).get(`/api/library/loans?status=${encodeURIComponent(status)}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'library/loans GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post('/api/library/loans', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'library/loans POST');
	}
};
