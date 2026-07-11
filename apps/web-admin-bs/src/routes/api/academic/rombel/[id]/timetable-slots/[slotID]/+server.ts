import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const slotID = requiredRouteParam(event.params.slotID, 'slotID');
		const data = await proxy(event).get(apiPath`/api/academic/rombel/${id}/timetable-slots/${slotID}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic rombel timetable slot GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const slotID = requiredRouteParam(event.params.slotID, 'slotID');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/academic/rombel/${id}/timetable-slots/${slotID}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic rombel timetable slot PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const slotID = requiredRouteParam(event.params.slotID, 'slotID');
		await proxy(event).del(apiPath`/api/academic/rombel/${id}/timetable-slots/${slotID}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'academic rombel timetable slot DELETE');
	}
};
