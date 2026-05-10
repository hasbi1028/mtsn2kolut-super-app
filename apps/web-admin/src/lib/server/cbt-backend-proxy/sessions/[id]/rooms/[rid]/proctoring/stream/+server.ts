import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, requiredRouteParam } from '$lib/server/api';
import { extractRoomDashboardEvents, proctoringEventStream } from '$lib/server/cbt-backend-proxy/proctoring-stream';

export const GET = (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const rid = requiredRouteParam(event.params.rid, 'rid');
		return proctoringEventStream(event, {
			path: cbtApiPath`/sessions/${id}/rooms/${rid}/proctoring`,
			extractEvents: extractRoomDashboardEvents,
			name: 'room-proctoring',
		});
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms/[rid]/proctoring/stream GET');
	}
};
