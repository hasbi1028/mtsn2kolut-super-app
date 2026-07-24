import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const items = await proxy(event).get<any[]>('/api/academic/curriculum/profiles');
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'curriculum profiles GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const result = await proxy(event).post<any>('/api/academic/curriculum/profiles', body);
		return json(result, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'curriculum profiles POST');
	}
};
