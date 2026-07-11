import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { apiPath, handleRouteError, proxy, requiredRouteParam, streamProxyResponse } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	if (!event.locals.user) {
		return json({ error: 'Unauthorized' }, { status: 401 });
	}

	try {
		const filename = requiredRouteParam(event.params.filename, 'filename');
		const response = await proxy(event).fetch(apiPath`/api/auth/account/avatar/${filename}`);
		return await streamProxyResponse(response, {
			fallbackMessage: 'Gagal mengambil foto profil.',
			defaultCacheControl: 'private, max-age=3600',
			headers: ['content-type', 'content-length', 'content-disposition', 'cache-control', 'x-content-type-options']
		});
	} catch (e) {
		return handleRouteError(e, 'auth/account/avatar/[filename] GET');
	}
};
