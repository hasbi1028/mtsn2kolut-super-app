import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const upstreamPath = apiPathWithQuery('/api/student-id-cards', event.url.searchParams);
		const data = await proxy(event).get(upstreamPath);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/kartu-siswa GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const base_url = typeof body.base_url === 'string' && body.base_url.trim()
			? body.base_url.trim()
			: event.url.origin;
		const data = await proxy(event).post('/api/student-id-cards/generate', { ...body, base_url });
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/kartu-siswa POST');
	}
};
