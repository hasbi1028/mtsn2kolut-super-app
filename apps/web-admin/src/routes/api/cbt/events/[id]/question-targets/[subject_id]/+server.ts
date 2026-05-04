import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const subjectId = requiredRouteParam(event.params.subject_id, 'subject_id');
		await proxy(event).del(apiPath`/api/cbt/events/${id}/question-targets/${subjectId}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/question-targets/[subject_id] DELETE');
	}
};
