import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get('/api/school-profile');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'school-profile GET');
	}
};

export const PUT: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).put('/api/school-profile', body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'school-profile PUT');
	}
};
