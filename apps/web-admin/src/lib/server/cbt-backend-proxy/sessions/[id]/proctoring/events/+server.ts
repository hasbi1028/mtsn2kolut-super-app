import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, apiPathWithQuery, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(
			apiPathWithQuery(cbtApiPath`/sessions/${id}/proctoring/events`, event.url.searchParams)
		);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/proctoring/events GET');
	}
};
