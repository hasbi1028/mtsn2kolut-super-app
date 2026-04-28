import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post(`/api/cbt/sessions/${event.params.id}/enroll`, body);
		return json(data, { status: 200 });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions enroll POST');
	}
};
