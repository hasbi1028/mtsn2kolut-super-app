import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const workerId = requiredRouteParam(
			(event.params as Record<string, string | undefined>).worker_id,
			'worker_id'
		);
		const body = (await event.request.json()) as { enabled?: boolean };
		const enabled = body.enabled === true;
		const data = await proxy(event).put<{ worker_id: string; enabled: boolean }>(
			`/api/pusaka/worker/enabled/${encodeURIComponent(workerId)}`,
			{ enabled }
		);
		return json({ data });
	} catch (e) {
		return handleRouteError(e, 'pusaka/worker/enabled');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const workerId = requiredRouteParam(
			(event.params as Record<string, string | undefined>).worker_id,
			'worker_id'
		);
		const data = await proxy(event).del<{ ok: boolean }>(
			`/api/pusaka/worker/enabled/${encodeURIComponent(workerId)}`
		);
		return json({ data });
	} catch (e) {
		return handleRouteError(e, 'pusaka/worker/enabled');
	}
};
