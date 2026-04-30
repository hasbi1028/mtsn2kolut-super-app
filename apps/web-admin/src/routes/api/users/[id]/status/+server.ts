import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const PATCH = async (event: RequestEvent) => {
	try {
		const body = await event.request.json() as Record<string, unknown>;
		const data = await proxy(event).patch(`/api/users/${event.params.id}/status`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'users/[id]/status PATCH');
	}
};
