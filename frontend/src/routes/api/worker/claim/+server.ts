import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { workerFetch, handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async ({ request }) => {
	const workerKey = request.headers.get('x-worker-key') ?? '';
	if (workerKey !== (process.env.WORKER_API_KEY ?? ''))
		return json({ error: 'Unauthorized' }, { status: 401 });

	try {
		const workerId = request.headers.get('x-worker-id') ?? 'unknown';
		const res = await workerFetch('POST', '/api/worker/claim', undefined, { worker_id: workerId });
		if (res.status === 204) return new Response(null, { status: 204 });

		const body = await res.json() as { data?: Record<string, unknown>; error?: string };
		if (!res.ok) return json({ error: body.error ?? 'upstream error' }, { status: res.status });
		return json({ job: body.data });
	} catch (e) {
		return handleRouteError(e, 'worker/claim');
	}
};
