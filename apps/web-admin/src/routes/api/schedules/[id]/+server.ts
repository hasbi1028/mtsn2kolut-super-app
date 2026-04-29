import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const body = await event.request.json();
		const sched = await proxy(event).put(`/api/schedules/${id}`, body);
		return json(sched);
	} catch (e) {
		return handleRouteError(e, 'schedules/[id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const result = await proxy(event).del(`/api/schedules/${id}`);
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'schedules/[id] DELETE');
	}
};
