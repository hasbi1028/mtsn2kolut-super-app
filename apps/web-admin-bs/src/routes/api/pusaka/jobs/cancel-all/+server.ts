import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const result = await proxy(event).post<{ ok: boolean; cancelled: number }>('/api/pusaka/jobs/cancel-all');
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'pusaka/jobs/cancel-all');
	}
};
