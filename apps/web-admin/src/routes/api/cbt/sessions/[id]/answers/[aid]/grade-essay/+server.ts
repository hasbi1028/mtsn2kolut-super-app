import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post(`/api/cbt/sessions/${event.params.id}/answers/${event.params.aid}/grade-essay`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/answers/[aid]/grade-essay POST');
	}
};
