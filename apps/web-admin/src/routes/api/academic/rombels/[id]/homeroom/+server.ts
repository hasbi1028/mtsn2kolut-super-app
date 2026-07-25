import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const result = await proxy(event).get<any>(`/api/academic/rombels/${id}/homeroom`);
		return new Response(JSON.stringify(result), {
			status: 200,
			headers: { 'content-type': 'application/json' },
		});
	} catch (e) {
		return handleRouteError(e, 'homeroom GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const body = await event.request.json();
		const result = await proxy(event).post<any>(`/api/academic/rombels/${id}/homeroom`, body);
		return new Response(JSON.stringify(result), {
			status: 200,
			headers: { 'content-type': 'application/json' },
		});
	} catch (e) {
		return handleRouteError(e, 'homeroom POST');
	}
};
