import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const PUT: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).put(`/api/governance/units/${event.params.id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'governance/units/[id] PUT');
	}
};

export const DELETE: RequestHandler = async (event) => {
	try {
		await proxy(event).del(`/api/governance/units/${event.params.id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'governance/units/[id] DELETE');
	}
};
