import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(`/api/tu/surat/disposisi/${event.params.id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat/disposisi/[id] GET');
	}
};

export const PUT: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).put(`/api/tu/surat/disposisi/${event.params.id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat/disposisi/[id] PUT');
	}
};

export const DELETE: RequestHandler = async (event) => {
	try {
		await proxy(event).del(`/api/tu/surat/disposisi/${event.params.id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'tu/surat/disposisi/[id] DELETE');
	}
};
