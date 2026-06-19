import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const assignmentID = requiredRouteParam(event.params.assignmentID, 'assignmentID');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/academic/rombel/${id}/homeroom-assignments/${assignmentID}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic rombel homeroom PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const assignmentID = requiredRouteParam(event.params.assignmentID, 'assignmentID');
		await proxy(event).del(apiPath`/api/academic/rombel/${id}/homeroom-assignments/${assignmentID}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'academic rombel homeroom DELETE');
	}
};
