import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const profileId = event.params.id;
		const level = event.url.searchParams.get('level') || '';
		const items = await proxy(event).get<any[]>(`/api/academic/curriculum/profiles/${profileId}/allocations?level=${level}`);
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'allocations GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const profileId = event.params.id;
		const body = await event.request.json();
		const result = await proxy(event).post<any>(`/api/academic/curriculum/profiles/${profileId}/allocations`, body);
		return json(result, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'allocations POST');
	}
};
