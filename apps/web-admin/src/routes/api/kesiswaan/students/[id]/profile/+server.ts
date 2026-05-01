import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const PUT: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).put(`/api/kesiswaan/students/${event.params.id}/profile`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/students/[id]/profile PUT');
	}
};
