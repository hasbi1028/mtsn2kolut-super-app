import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const PATCH = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).patch(cbtBackendPath('/questions/bulk-workflow'), body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/bulk-workflow PATCH');
	}
};
