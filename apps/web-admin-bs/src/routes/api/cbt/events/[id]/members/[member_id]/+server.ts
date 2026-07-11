import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const memberID = requiredRouteParam(event.params.member_id, 'member_id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/cbt/events/${id}/members/${memberID}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/members/[member_id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const memberID = requiredRouteParam(event.params.member_id, 'member_id');
		await proxy(event).del(apiPath`/api/cbt/events/${id}/members/${memberID}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/members/[member_id] DELETE');
	}
};
