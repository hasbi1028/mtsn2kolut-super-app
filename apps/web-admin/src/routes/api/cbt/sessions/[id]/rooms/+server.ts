import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(`/api/cbt/sessions/${event.params.id}/rooms`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post(`/api/cbt/sessions/${event.params.id}/rooms`, body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms POST');
	}
};
