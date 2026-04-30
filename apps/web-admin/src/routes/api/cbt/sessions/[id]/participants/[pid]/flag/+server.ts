import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post(`/api/cbt/sessions/${event.params.id}/participants/${event.params.pid}/flag`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/participants/[pid]/flag POST');
	}
};
