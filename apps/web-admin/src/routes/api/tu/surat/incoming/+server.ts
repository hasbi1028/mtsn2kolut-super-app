import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const search = event.url.searchParams.get('search') ?? '';
		const status = event.url.searchParams.get('status') ?? '';
		const data = await proxy(event).get(
			`/api/tu/surat/incoming?search=${encodeURIComponent(search)}&status=${encodeURIComponent(status)}`
		);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat/incoming GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post('/api/tu/surat/incoming', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'tu/surat/incoming POST');
	}
};
