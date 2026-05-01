import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const PATCH: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).patch(`/api/tu/surat/incoming/${event.params.id}/status`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat/incoming/[id]/status PATCH');
	}
};
