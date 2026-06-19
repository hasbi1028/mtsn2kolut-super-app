import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/public/student-cards/portal-login/start', body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'public student card portal-login start');
	}
};
