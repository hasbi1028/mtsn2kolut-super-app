import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request).catch(() => ({}));
		const data = await proxy(event).post(cbtApiPath`/approvals/${id}/revoke`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/approvals/[id]/revoke POST');
	}
};
