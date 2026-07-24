import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const body = await event.request.json();
		body.id = id;
		const result = await proxy(event).put<any>(`/api/academic/timetable/slots/${id}`, body);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'timetable slots PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		await proxy(event).del(`/api/academic/timetable/slots/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'timetable slots DELETE');
	}
};
