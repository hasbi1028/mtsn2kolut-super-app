import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const item = await proxy(event).put(`/api/website/content/${event.params.id}`, body);
		return json(item);
	} catch (e) {
		return handleRouteError(e, 'website/content/:id PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		await proxy(event).del(`/api/website/content/${event.params.id}`);
		return json({ success: true });
	} catch (e) {
		return handleRouteError(e, 'website/content/:id DELETE');
	}
};
