import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const res = await proxy(event).get('/api/rbac/matrix');
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'rbac matrix GET');
	}
};
