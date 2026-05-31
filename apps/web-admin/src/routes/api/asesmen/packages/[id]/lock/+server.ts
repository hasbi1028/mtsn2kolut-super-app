import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readOptionalRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	const id = requiredRouteParam(event.params.id, 'id');
	try {
		const body = await readOptionalRequestJson<Record<string, unknown>>(event.request, {});
		const data = await proxy(event).post(apiPath`/api/asesmen/packages/${id}/lock`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/packages/[id]/lock POST');
	}
};
