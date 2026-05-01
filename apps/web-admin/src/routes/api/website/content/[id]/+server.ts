import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const item = await proxy(event).put(apiPath`/api/website/content/${id}`, body);
		return json(item);
	} catch (e) {
		return handleRouteError(e, 'website/content/:id PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(apiPath`/api/website/content/${id}`);
		return json({ success: true });
	} catch (e) {
		return handleRouteError(e, 'website/content/:id DELETE');
	}
};
