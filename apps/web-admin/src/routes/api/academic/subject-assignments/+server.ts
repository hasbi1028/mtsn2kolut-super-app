import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const params = event.url.searchParams.toString();
		const result = await proxy(event).get<any>(`/api/academic/subject-assignments?${params}`);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'subject assignments GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const result = await proxy(event).post<any>('/api/academic/subject-assignments', body);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'subject assignments POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		await proxy(event).del(`/api/academic/subject-assignments/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'subject assignments DELETE');
	}
};
