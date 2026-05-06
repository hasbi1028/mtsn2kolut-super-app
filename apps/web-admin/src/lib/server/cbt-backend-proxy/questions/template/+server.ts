import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, streamProxyResponse } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const res = await proxy(event).fetch(cbtBackendPath('/questions/template'));
		return await streamProxyResponse(res, {
			defaultContentType: 'text/csv; charset=utf-8',
			defaultCacheControl: 'no-store',
		});
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/template GET');
	}
};
