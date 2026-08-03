import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

type WorkerCap = {
	worker_id: string;
	cap: number;
};

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get<{ caps: WorkerCap[] }>('/api/pusaka/worker/caps');
		return json({ data: { caps: data.caps ?? [] } });
	} catch (e) {
		return handleRouteError(e, 'pusaka/worker/caps');
	}
};
