import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const res = await proxy(event).post('/api/users/generate-from-employees', {});
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'users generate-from-employees POST');
	}
};
