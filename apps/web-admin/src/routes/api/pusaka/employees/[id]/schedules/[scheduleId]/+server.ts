import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const DELETE = async (event: RequestEvent) => {
	try {
		const { id, scheduleId } = event.params;
		const result = await proxy(event).del(`/api/pusaka/employees/${id}/schedules/${scheduleId}`);
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'pusaka/employees/[id]/schedules/[scheduleId] DELETE');
	}
};
