import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, requiredRouteParam, streamProxyResponse } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const filename = requiredRouteParam(event.params.filename, 'filename');
		const response = await proxy(event).fetch(apiPath`/api/kesiswaan/student-photos/${filename}`);
		return await streamProxyResponse(response, {
			fallbackMessage: 'Gagal mengambil foto siswa.',
			defaultCacheControl: 'private, max-age=3600',
			headers: ['content-type', 'cache-control']
		});
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/student-photos/[filename] GET');
	}
};
