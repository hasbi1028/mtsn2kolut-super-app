import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json() as Record<string, unknown>;
		const data = await proxy(event).post(`/api/parents/${event.params.id}/unlink`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'parents/[id]/unlink POST');
	}
};
