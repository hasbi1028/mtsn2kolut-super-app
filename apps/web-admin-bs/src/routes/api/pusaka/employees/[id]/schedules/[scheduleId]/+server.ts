import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const scheduleId = requiredRouteParam(event.params.scheduleId, 'scheduleId');
		const result = await proxy(event).del(apiPath`/api/pusaka/employees/${id}/schedules/${scheduleId}`);
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'pusaka/employees/[id]/schedules/[scheduleId] DELETE');
	}
};
