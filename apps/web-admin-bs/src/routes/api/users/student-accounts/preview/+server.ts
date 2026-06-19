import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const res = await proxy(event).get('/api/users/student-accounts/preview');
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'users/student-accounts/preview GET');
	}
};
