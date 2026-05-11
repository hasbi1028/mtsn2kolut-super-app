import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/academic');
		const subjects = (data as { subjects?: unknown }).subjects ?? [];
		return json(subjects);
	} catch (e) {
		return handleRouteError(e, 'academic subjects GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/academic/subjects', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'academic subjects POST');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/academic/subjects/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic subjects PUT');
	}
};
