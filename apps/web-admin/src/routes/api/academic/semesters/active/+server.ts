import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const result = await proxy(event).get<any>('/api/academic/semesters/active');
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'active semester GET');
	}
};
