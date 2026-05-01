import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, requiredRouteParam } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/tu/surat-keterangan/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat-keterangan/[id] GET');
	}
};
