import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

type WebFallbackPolicyRequest = {
	allow_web_fallback?: boolean;
	reason?: string;
};

export const PATCH = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const rid = requiredRouteParam(event.params.rid, 'rid');
		const body = await readRequestJson<WebFallbackPolicyRequest>(event.request, 8 * 1024);
		const data = await proxy(event).patch(cbtApiPath`/sessions/${id}/rooms/${rid}/web-fallback`, {
			allow_web_fallback: body.allow_web_fallback === true,
			reason: typeof body.reason === 'string' ? body.reason.trim() : '',
		});
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms/[rid]/web-fallback PATCH');
	}
};
