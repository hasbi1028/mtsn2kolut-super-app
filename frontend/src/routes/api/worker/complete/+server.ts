import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { workerFetch } from '$lib/server/api';

interface CompleteBody {
	job_id:    string;
	tanggal:   string;
	jam_masuk:  string;
	jam_pulang: string;
}

export const POST: RequestHandler = async ({ request }) => {
	const workerKey = request.headers.get('x-worker-key') ?? '';
	if (workerKey !== (process.env.WORKER_API_KEY ?? ''))
		return json({ error: 'Unauthorized' }, { status: 401 });

	const { job_id, tanggal, jam_masuk, jam_pulang } = await request.json() as CompleteBody;
	if (!job_id) return json({ error: 'job_id wajib' }, { status: 400 });

	const res = await workerFetch('POST', `/api/worker/jobs/${job_id}/complete`, undefined, {
		tanggal, jam_masuk, jam_pulang,
	});

	if (res.ok || res.status === 204) return json({ ok: true });
	const body = await res.json() as { error?: string };
	return json({ error: body.error ?? 'upstream error' }, { status: res.status });
};
