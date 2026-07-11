import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, apiPathWithQuery, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPathWithQuery(apiPath`/api/student-id-cards/${id}/events`, event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/kartu-siswa/[id]/events GET');
	}
};
