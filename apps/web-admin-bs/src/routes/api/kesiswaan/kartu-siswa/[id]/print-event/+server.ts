import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readOptionalRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readOptionalRequestJson<Record<string, unknown>>(event.request, {});
		const data = await proxy(event).post(apiPath`/api/student-id-cards/${id}/print-event`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/kartu-siswa/[id]/print-event POST');
	}
};
