import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PATCH: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).patch(apiPath`/api/tu/surat/incoming/${id}/status`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat/incoming/[id]/status PATCH');
	}
};
