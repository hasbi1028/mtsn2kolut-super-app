import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const memberId = requiredRouteParam(event.params.member_id, 'member_id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(cbtApiPath`/events/${id}/members/${memberId}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/members/[member_id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const memberId = requiredRouteParam(event.params.member_id, 'member_id');
		await proxy(event).del(cbtApiPath`/events/${id}/members/${memberId}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/members/[member_id] DELETE');
	}
};
