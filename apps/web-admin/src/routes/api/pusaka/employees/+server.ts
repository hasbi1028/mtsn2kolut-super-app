import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const items = await proxy(event).get('/api/employees?with_status=1&scope=pusaka');
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'pusaka/employees GET');
	}
};
