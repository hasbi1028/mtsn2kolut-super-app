import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const params = new URLSearchParams();
		const search = event.url.searchParams.get('search') ?? '';
		const activeOnly = event.url.searchParams.get('active_only') ?? '';
		if (search) params.set('search', search);
		if (activeOnly) params.set('active_only', activeOnly);
		const suffix = params.toString() ? `?${params.toString()}` : '';
		const data = await proxy(event).get(`/api/tu/archives/categories${suffix}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/archives/categories GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post('/api/tu/archives/categories', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'tu/archives/categories POST');
	}
};
