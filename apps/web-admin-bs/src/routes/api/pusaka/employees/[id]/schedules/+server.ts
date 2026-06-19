import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const schedules = await proxy(event).get(apiPath`/api/pusaka/employees/${id}/schedules`);
		return json(schedules);
	} catch (e) {
		return handleRouteError(e, 'pusaka/employees/[id]/schedules GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const sched = await proxy(event).post(apiPath`/api/pusaka/employees/${id}/schedules`, body);
		return json(sched);
	} catch (e) {
		return handleRouteError(e, 'pusaka/employees/[id]/schedules POST');
	}
};
