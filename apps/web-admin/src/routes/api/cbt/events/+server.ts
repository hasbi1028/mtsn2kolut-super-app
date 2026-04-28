import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/cbt/events');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post('/api/cbt/events', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/events POST');
	}
};
