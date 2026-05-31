import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	const id = requiredRouteParam(event.params.id, 'id');
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/asesmen/packages/${id}/questions`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/packages/[id]/questions PUT');
	}
};
