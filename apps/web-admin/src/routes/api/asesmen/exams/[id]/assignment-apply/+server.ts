import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request, 32 << 10);
		const data = await proxy(event).post(apiPath`/api/asesmen/exams/${id}/assignment-apply`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams/[id]/assignment-apply POST');
	}
};
