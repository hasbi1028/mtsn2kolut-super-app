import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const body = await event.request.json();
		const result = await proxy(event).put<any>(`/api/kesiswaan/murid/${id}/profile`, body);
		return new Response(JSON.stringify(result), { status: 200, headers: { 'content-type': 'application/json' } });
	} catch (e) {
		return handleRouteError(e, 'kesiswaan murid profile PUT');
	}
};
