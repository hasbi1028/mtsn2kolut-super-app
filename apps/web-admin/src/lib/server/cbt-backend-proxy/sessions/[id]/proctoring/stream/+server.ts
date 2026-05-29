import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, requiredRouteParam } from '$lib/server/api';
import { extractSessionEvents, proctoringEventStream } from '$lib/server/cbt-backend-proxy/proctoring-stream';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		return proctoringEventStream(event, {
			name: 'session-proctoring-events',
			path: apiPathWithQuery(cbtApiPath`/sessions/${id}/proctoring/events`, new URLSearchParams({ limit: '100' })),
			extractEvents: extractSessionEvents,
		});
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/proctoring/stream GET');
	}
};
