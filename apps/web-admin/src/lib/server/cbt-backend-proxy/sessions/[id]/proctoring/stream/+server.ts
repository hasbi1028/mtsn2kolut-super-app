import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, requiredRouteParam } from '$lib/server/api';
import { extractSessionEvents, proctoringEventStream } from '$lib/server/cbt-backend-proxy/proctoring-stream';

export const GET = (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const params = new URLSearchParams(event.url.searchParams);
		if (!params.has('limit')) params.set('limit', '100');
		return proctoringEventStream(event, {
			path: apiPathWithQuery(cbtApiPath`/sessions/${id}/proctoring/events`, params),
			extractEvents: extractSessionEvents,
			name: 'session-proctoring',
		});
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/proctoring/stream GET');
	}
};
