import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readOptionalRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readOptionalRequestJson<Record<string, unknown>>(event.request, {});
		const data = await proxy(event).post(apiPath`/api/tu/surat-keterangan/${id}/cancel`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat-keterangan/[id]/cancel POST');
	}
};
