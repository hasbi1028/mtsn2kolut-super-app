import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(cbtApiPath`/sessions/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id] GET');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(cbtApiPath`/sessions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id] DELETE');
	}
};
