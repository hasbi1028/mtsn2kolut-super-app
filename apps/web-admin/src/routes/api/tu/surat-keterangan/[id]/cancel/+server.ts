import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json().catch(() => ({}));
		const data = await proxy(event).post(`/api/tu/surat-keterangan/${event.params.id}/cancel`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat-keterangan/[id]/cancel POST');
	}
};
