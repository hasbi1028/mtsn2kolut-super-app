import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const result = await proxy(event).get<any>(`/api/kesiswaan/murid/${id}`);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'murid profile GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const body = await event.request.json();
		const result = await proxy(event).put<any>(`/api/kesiswaan/murid/${id}/profile`, body);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'murid profile PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const result = await proxy(event).del<any>(`/api/kesiswaan/murid/${id}`);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'murid delete DELETE');
	}
};
