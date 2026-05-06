import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(cbtBackendPath('/events'));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(cbtBackendPath('/events'), body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/events POST');
	}
};
