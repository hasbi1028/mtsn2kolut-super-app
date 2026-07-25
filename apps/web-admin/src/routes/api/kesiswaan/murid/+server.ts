import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const params = event.url.searchParams.toString();
		const result = await proxy(event).get<any>(`/api/kesiswaan/murid?${params}`);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'murid list GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const result = await proxy(event).post<any>('/api/kesiswaan/murid', body);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'murid create POST');
	}
};
