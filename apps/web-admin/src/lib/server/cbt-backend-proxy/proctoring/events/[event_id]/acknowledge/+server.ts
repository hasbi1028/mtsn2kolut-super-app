import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const eventID = requiredRouteParam(event.params.event_id, 'event_id');
		const body = await readRequestJson(event.request);
		const data = await proxy(event).post(cbtApiPath`/proctoring/events/${eventID}/acknowledge`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/proctoring/events/[event_id]/acknowledge POST');
	}
};
