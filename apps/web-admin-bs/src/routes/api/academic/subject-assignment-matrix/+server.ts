import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/academic/subject-assignment-matrix');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic subject assignment matrix GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put('/api/academic/subject-assignment-matrix', body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic subject assignment matrix PUT');
	}
};
