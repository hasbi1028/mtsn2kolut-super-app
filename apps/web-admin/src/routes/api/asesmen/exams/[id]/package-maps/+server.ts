import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/asesmen/exams/${id}/package-maps`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams/[id]/package-maps GET');
	}
};

export const PUT: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/asesmen/exams/${id}/package-maps`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams/[id]/package-maps PUT');
	}
};
