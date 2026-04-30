import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const schedules = await proxy(event).get(`/api/pusaka/employees/${id}/schedules`);
		return json(schedules);
	} catch (e) {
		return handleRouteError(e, 'pusaka/employees/[id]/schedules GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const body = await event.request.json();
		const sched = await proxy(event).post(`/api/pusaka/employees/${id}/schedules`, body);
		return json(sched);
	} catch (e) {
		return handleRouteError(e, 'pusaka/employees/[id]/schedules POST');
	}
};
