import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const PUT: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).put(`/api/governance/assignments/${event.params.id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'governance/assignments/[id] PUT');
	}
};

export const DELETE: RequestHandler = async (event) => {
	try {
		await proxy(event).del(`/api/governance/assignments/${event.params.id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'governance/assignments/[id] DELETE');
	}
};
