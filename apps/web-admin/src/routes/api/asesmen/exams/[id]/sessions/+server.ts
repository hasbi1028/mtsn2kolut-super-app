import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/cbt/events/${id}/sessions`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams/[id]/sessions GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request, 32 << 10);
		const data = await proxy(event).post(apiPath`/api/cbt/sessions`, {
			...body,
			event_id: id
		});
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams/[id]/sessions POST');
	}
};
