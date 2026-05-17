import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const rid = requiredRouteParam(event.params.rid, 'rid');
		const data = await proxy(event).get(cbtApiPath`/sessions/${id}/rooms/${rid}/proctoring/report-data`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms/[rid]/proctoring/report-data GET');
	}
};
