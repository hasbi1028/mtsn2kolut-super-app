import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const res = await proxy(event).get('/api/users/generate-from-employees/preview');
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'users generate-from-employees preview GET');
	}
};
