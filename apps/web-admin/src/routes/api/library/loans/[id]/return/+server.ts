import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).post(`/api/library/loans/${event.params.id}/return`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'library/loans return');
	}
};
