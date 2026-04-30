import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { proxy, ApiError, handleRouteError } from '$lib/server/api';

export const DELETE: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		await proxy(event).del(`/api/auth/sessions/${event.params.id}`);
		return json({ ok: true });
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/sessions DELETE');
	}
};
