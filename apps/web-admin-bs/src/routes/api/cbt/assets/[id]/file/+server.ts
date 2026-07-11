import { apiPath, handleRouteError, proxy, requiredRouteParam, streamProxyResponse } from '$lib/server/api';
import type { RequestEvent } from './$types';

export async function GET(event: RequestEvent) {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const response = await proxy(event).fetch(apiPath`/api/cbt/assets/${id}/file`);
		return streamProxyResponse(response, { fallbackMessage: 'Berkas aset soal tidak tersedia' });
	} catch (e) {
		return handleRouteError(e, 'cbt/assets/file GET');
	}
}
