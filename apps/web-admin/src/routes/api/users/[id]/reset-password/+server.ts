import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const res = await proxy(event).post(apiPath`/api/users/${id}/reset-password`, body);
		return json(res ?? { ok: true });
	} catch (e) {
		return handleRouteError(e, 'users reset-password POST');
	}
};
