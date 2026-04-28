import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const PATCH = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).patch(`/api/cbt/events/${event.params.id}/status`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/status PATCH');
	}
};
