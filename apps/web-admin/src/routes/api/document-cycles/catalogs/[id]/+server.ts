import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const PUT: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).put(`/api/document-cycles/catalogs/${event.params.id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'document-cycles/catalogs/[id] PUT');
	}
};

export const DELETE: RequestHandler = async (event) => {
	try {
		await proxy(event).del(`/api/document-cycles/catalogs/${event.params.id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'document-cycles/catalogs/[id] DELETE');
	}
};
