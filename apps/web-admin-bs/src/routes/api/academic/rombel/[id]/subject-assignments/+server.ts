import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/academic/rombel/${id}/subject-assignments`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic rombel subject assignments GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(apiPath`/api/academic/rombel/${id}/subject-assignments`, body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'academic rombel subject assignments POST');
	}
};
