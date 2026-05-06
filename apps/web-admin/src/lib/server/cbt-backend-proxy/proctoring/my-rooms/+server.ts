import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(cbtBackendPath('/proctoring/my-rooms'));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/proctoring/my-rooms GET');
	}
};
