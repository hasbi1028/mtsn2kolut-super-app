import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/asesmen/exams/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams/[id] GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request, 32 << 10);
		const data = await proxy(event).put(apiPath`/api/asesmen/exams/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams/[id] PUT');
	}
};
