import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, requiredRouteParam, streamProxyResponse } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const res = await proxy(event).fetch(cbtApiPath`/sessions/${id}/proctoring/stream`, {
			headers: { Accept: 'text/event-stream' },
		});
		return streamProxyResponse(res, { defaultContentType: 'text/event-stream; charset=utf-8' });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/proctoring/stream GET');
	}
};
