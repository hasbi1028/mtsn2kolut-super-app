import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { proxy, ApiError, handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		await proxy(event).post('/api/auth/logout-all', {});
		event.cookies.delete('access_token', { path: '/' });
		event.cookies.delete('refresh_token', { path: '/' });
		return json({ ok: true });
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/logout-all');
	}
};
