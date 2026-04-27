import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { workerFetch, handleRouteError } from '$lib/server/api';

interface FailBody { job_id: string; error: string }

export const POST: RequestHandler = async ({ request }) => {
	const workerKey = request.headers.get('x-worker-key') ?? '';
	if (workerKey !== (process.env.WORKER_API_KEY ?? ''))
		return json({ error: 'Unauthorized' }, { status: 401 });

	try {
		const { job_id, error } = await request.json() as FailBody;
		if (!job_id) return json({ error: 'job_id wajib' }, { status: 400 });

		const res = await workerFetch('POST', `/api/worker/jobs/${job_id}/fail`, undefined, { error });
		if (res.ok || res.status === 204) return json({ ok: true });
		const body = await res.json() as { error?: string };
		return json({ error: body.error ?? 'upstream error' }, { status: res.status });
	} catch (e) {
		return handleRouteError(e, 'worker/fail');
	}
};
