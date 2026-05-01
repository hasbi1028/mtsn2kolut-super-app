import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, requiredRouteParam, streamProxyResponse } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const response = await proxy(event).fetch(apiPath`/api/tu/archives/documents/${id}/file`);
		return await streamProxyResponse(response, {
			fallbackMessage: 'Gagal mengambil file arsip.'
		});
	} catch (e) {
		return handleRouteError(e, 'tu/archives/documents/[id]/file GET');
	}
};
