import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { rawDb } from '$lib/server/db';
import { checkWorkerAuth } from '$lib/server/workerAuth';

const RETRY_BASE_SECONDS = 30;

interface FailBody {
	job_id: string;
	error:  string;
}

export const POST: RequestHandler = async ({ request }) => {
	if (!checkWorkerAuth(request)) return json({ error: 'Unauthorized' }, { status: 401 });

	const { job_id, error } = await request.json() as FailBody;
	if (!job_id) return json({ error: 'job_id wajib' }, { status: 400 });

	const job = rawDb.prepare('SELECT attempts, max_attempts FROM jobs WHERE id = ?')
		.get(job_id) as { attempts: number; max_attempts: number } | undefined;
	if (!job) return json({ error: 'Job tidak ditemukan' }, { status: 404 });

	const nextAttempts = Number(job.attempts) + 1;
	const maxAttempts  = Number(job.max_attempts);

	if (nextAttempts < maxAttempts) {
		const delaySeconds = RETRY_BASE_SECONDS * Math.pow(2, nextAttempts - 1);
		rawDb.prepare(`
			UPDATE jobs
			SET status='queued', attempts=?, error_message=?,
			    next_retry_at=datetime(CURRENT_TIMESTAMP, '+' || ? || ' seconds'),
			    updated_at=CURRENT_TIMESTAMP
			WHERE id = ?
		`).run(nextAttempts, error, delaySeconds, job_id);
		return json({ ok: true, action: 'retry', attempts: nextAttempts, retry_in_seconds: delaySeconds });
	}

	rawDb.prepare(`
		UPDATE jobs
		SET status='failed', attempts=?, error_message=?, next_retry_at=NULL, updated_at=CURRENT_TIMESTAMP
		WHERE id = ?
	`).run(nextAttempts, error, job_id);
	return json({ ok: true, action: 'failed', attempts: nextAttempts });
};
