import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const search = event.url.searchParams.get('search') ?? '';
		const activeOnly = event.url.searchParams.get('active_only') ?? '';
		const data = await proxy(event).get(
			`/api/governance/assignments?search=${encodeURIComponent(search)}&active_only=${encodeURIComponent(activeOnly)}`
		);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'governance/assignments GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post('/api/governance/assignments', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'governance/assignments POST');
	}
};
