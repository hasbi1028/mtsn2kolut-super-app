import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const code = requiredRouteParam(event.params.code, 'code');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const res = await proxy(event).put(apiPath`/api/rbac/permissions/${code}`, body);
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'rbac permission PUT');
	}
};
