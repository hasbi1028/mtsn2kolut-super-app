import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	const id = event.params.id ?? '';
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(cbtApiPath`/packages/${id}/questions`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/packages/[id]/questions PUT');
	}
};
