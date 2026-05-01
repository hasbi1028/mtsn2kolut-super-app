import { json } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export async function GET(event) {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}
	try {
		const data = await proxy(event).get('/api/auth/preferences/sidebar');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'auth/preferences/sidebar GET');
	}
}

export async function PATCH(event) {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}
	try {
		const body = await event.request.json();
		const data = await proxy(event).patch('/api/auth/preferences/sidebar', body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'auth/preferences/sidebar PATCH');
	}
}
