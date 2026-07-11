import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const base_url = typeof body.base_url === 'string' && body.base_url.trim()
			? body.base_url.trim()
			: event.url.origin;
		const data = await proxy(event).post(apiPath`/api/student-id-cards/${id}/reissue`, { ...body, base_url });
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/kartu-siswa/[id]/reissue POST');
	}
};
