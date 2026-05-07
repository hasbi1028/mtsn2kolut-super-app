import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { apiPath, proxy, ApiError, handleRouteError, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const DELETE: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(apiPath`/api/auth/sessions/${id}`);
		// If the user just revoked their own active session, drop the cookies
		// immediately so this browser sees the logout state without waiting for
		// the access token to expire.
		if (event.locals.user.session_id && event.locals.user.session_id === id) {
			event.cookies.delete('access_token', { path: '/' });
			event.cookies.delete('refresh_token', { path: '/' });
		}
		return json({ ok: true });
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/sessions DELETE');
	}
};

export const PATCH: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		await proxy(event).patch(apiPath`/api/auth/sessions/${id}`, body);
		return json({ ok: true });
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/sessions PATCH');
	}
};
