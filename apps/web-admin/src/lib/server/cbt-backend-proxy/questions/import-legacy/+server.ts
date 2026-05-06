import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, jsonProxyResponse, proxy } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const form = await event.request.formData();
		const res = await proxy(event).fetch(cbtBackendPath('/questions/import-legacy'), {
			method: 'POST',
			body: form,
		});
		return await jsonProxyResponse<{ data?: unknown }, unknown>(res, {
			map: (data) => data.data ?? data,
		});
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/import-legacy POST');
	}
};
