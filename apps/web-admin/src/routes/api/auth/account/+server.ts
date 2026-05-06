import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { proxy, ApiError, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const data = await proxy(event).get('/api/auth/account');
		return json(data);
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/account GET');
	}
};
