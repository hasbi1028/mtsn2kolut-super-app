import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError, readRequestJson } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const res = await proxy(event).get('/api/users');
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'users GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const res = await proxy(event).post('/api/users', body);
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'users POST');
	}
};
