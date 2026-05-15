import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readOptionalRequestJson } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	const id = event.params.id ?? '';
	try {
		const body = await readOptionalRequestJson<Record<string, unknown>>(event.request, {});
		const data = await proxy(event).post(cbtApiPath`/packages/${id}/lock`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/packages/[id]/lock POST');
	}
};
