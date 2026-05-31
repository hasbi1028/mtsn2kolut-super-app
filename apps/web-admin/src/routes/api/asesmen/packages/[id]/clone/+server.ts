import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	const id = requiredRouteParam(event.params.id, 'id');
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(apiPath`/api/asesmen/packages/${id}/clone`, body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'asesmen/packages/[id]/clone POST');
	}
};
