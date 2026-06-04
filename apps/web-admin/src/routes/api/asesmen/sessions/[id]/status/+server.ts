import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

const VALID_SESSION_STATUSES = new Set(['draft', 'scheduled', 'active', 'finished', 'cancelled']);

export const PATCH = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request, 8 << 10);
		const status = typeof body.status === 'string' ? body.status.trim() : '';
		if (!VALID_SESSION_STATUSES.has(status)) {
			throw new ApiError(400, 'Status sesi CBT tidak valid');
		}
		const data = await proxy(event).patch(apiPath`/api/asesmen/sessions/${id}/status`, { status });
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/sessions/[id]/status PATCH');
	}
};
