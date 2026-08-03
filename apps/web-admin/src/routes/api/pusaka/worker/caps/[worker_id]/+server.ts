import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const workerId = requiredRouteParam(
			(event.params as Record<string, string | undefined>).worker_id,
			'worker_id'
		);
		const body = (await event.request.json()) as { cap?: number };
		const cap = Number(body.cap);
		if (!Number.isInteger(cap) || cap < 1 || cap > 35) {
			return json({ error: 'cap harus bilangan bulat antara 1 dan 35' }, { status: 400 });
		}
		const data = await proxy(event).put<{ worker_id: string; cap: number }>(
			`/api/pusaka/worker/caps/${encodeURIComponent(workerId)}`,
			{ cap }
		);
		return json({ data });
	} catch (e) {
		return handleRouteError(e, 'pusaka/worker/caps');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const workerId = requiredRouteParam(
			(event.params as Record<string, string | undefined>).worker_id,
			'worker_id'
		);
		const data = await proxy(event).del<{ ok: boolean }>(
			`/api/pusaka/worker/caps/${encodeURIComponent(workerId)}`
		);
		return json({ data });
	} catch (e) {
		return handleRouteError(e, 'pusaka/worker/caps');
	}
};
