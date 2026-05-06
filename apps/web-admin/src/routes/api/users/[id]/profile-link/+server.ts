import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PATCH = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const res = await proxy(event).patch(apiPath`/api/users/${id}/profile-link`, body);
		return json(res ?? { ok: true });
	} catch (e) {
		return handleRouteError(e, 'users profile-link PATCH');
	}
};
