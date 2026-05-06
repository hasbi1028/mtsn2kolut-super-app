import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { proxy, ApiError, apiPathWithQuery, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const params = new URLSearchParams(event.url.searchParams);
		params.delete('user_id');
		const data = await proxy(event).get(apiPathWithQuery('/api/auth/account/change-history', params));
		return json(data);
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/account/change-history GET');
	}
};
