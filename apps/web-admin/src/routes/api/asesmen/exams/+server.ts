import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/asesmen/exams', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request, 32 << 10);
		const data = await proxy(event).post('/api/asesmen/exams', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams POST');
	}
};
