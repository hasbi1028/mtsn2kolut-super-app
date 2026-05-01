import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const PUT: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).put(`/api/kesiswaan/violations/${event.params.id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/violations/[id] PUT');
	}
};

export const DELETE: RequestHandler = async (event) => {
	try {
		await proxy(event).del(`/api/kesiswaan/violations/${event.params.id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/violations/[id] DELETE');
	}
};
