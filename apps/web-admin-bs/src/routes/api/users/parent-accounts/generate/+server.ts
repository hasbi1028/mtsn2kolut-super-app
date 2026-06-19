import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const res = await proxy(event).post('/api/users/parent-accounts/generate', {});
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'users/parent-accounts/generate POST');
	}
};
