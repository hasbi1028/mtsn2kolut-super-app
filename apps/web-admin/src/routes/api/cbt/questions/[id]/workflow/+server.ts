import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const PATCH = async (event: RequestEvent) => {
	try {
		const { id } = event.params;
		const body = await event.request.json();
		const data = await proxy(event).patch(`/api/cbt/questions/${id}/workflow`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/[id]/workflow PATCH');
	}
};
