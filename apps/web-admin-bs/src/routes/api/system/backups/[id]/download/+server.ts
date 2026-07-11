import type { RequestHandler } from './$types';
import { apiPath, handleRouteError, proxy, requiredRouteParam, streamProxyResponse } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const response = await proxy(event).fetch(apiPath`/api/system/backups/${id}/download`);
		return await streamProxyResponse(response, {
			fallbackMessage: 'Gagal mengunduh file backup.',
			defaultContentType: 'application/octet-stream',
			headers: ['content-type', 'content-length', 'content-disposition', 'cache-control', 'x-content-type-options']
		});
	} catch (error) {
		return handleRouteError(error, 'system/backups/[id]/download GET');
	}
};
