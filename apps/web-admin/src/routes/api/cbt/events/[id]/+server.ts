import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(`/api/cbt/events/${event.params.id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id] GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).put(`/api/cbt/events/${event.params.id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		await proxy(event).del(`/api/cbt/events/${event.params.id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id] DELETE');
	}
};
