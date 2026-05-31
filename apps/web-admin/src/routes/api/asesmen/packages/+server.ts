import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/asesmen/packages', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/packages GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/asesmen/packages', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'asesmen/packages POST');
	}
};
