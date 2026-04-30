import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(`/api/cbt/sessions/${event.params.id}/ungraded-essays`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/ungraded-essays GET');
	}
};
