import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const result = await proxy(event).get<any>(`/api/class-journal/sessions/${event.params.id}`);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'journal session GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const result = await proxy(event).put<any>(`/api/class-journal/sessions/${event.params.id}`, body);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'journal session PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const result = await proxy(event).del<any>(`/api/class-journal/sessions/${event.params.id}`);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'journal session DELETE');
	}
};
