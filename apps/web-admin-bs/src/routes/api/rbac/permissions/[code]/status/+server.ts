import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PATCH = async (event: RequestEvent) => {
	try {
		const code = requiredRouteParam(event.params.code, 'code');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const res = await proxy(event).patch(apiPath`/api/rbac/permissions/${code}/status`, body);
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'rbac permission status PATCH');
	}
};
