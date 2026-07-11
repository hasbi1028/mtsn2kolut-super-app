import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { proxy, ApiError, apiPath, handleRouteError, requiredRouteParam } from '$lib/server/api';

async function cancelOwnRequest(event: RequestEvent) {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).post(apiPath`/api/auth/account/change-requests/${id}/cancel`, {});
		return json(data);
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			return json({ error: 'Unauthorized' }, { status: 401 });
		}
		return handleRouteError(e, 'auth/account/change-requests/:id/cancel');
	}
}

export const POST: RequestHandler = cancelOwnRequest;
export const PATCH: RequestHandler = cancelOwnRequest;
