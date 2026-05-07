import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/portal/parent/children');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'portal/orang-tua/children GET');
	}
};
