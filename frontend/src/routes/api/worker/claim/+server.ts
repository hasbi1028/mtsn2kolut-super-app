import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { rawDb } from '$lib/server/db';
import { checkWorkerAuth } from '$lib/server/workerAuth';

export const POST: RequestHandler = ({ request }) => {
	if (!checkWorkerAuth(request)) return json({ error: 'Unauthorized' }, { status: 401 });

	const workerId = request.headers.get('x-worker-id') ?? 'unknown';

	const job = rawDb.transaction(() => {
		const row = rawDb.prepare(`
			SELECT j.id, j.employee_id, j.run_type, j.attempts, j.max_attempts,
			       e.pusaka_username, e.pusaka_password
			FROM jobs j
			JOIN employees e ON e.id = j.employee_id
			WHERE j.status = 'queued'
			  AND (j.next_retry_at IS NULL OR datetime(j.next_retry_at) <= CURRENT_TIMESTAMP)
			ORDER BY j.created_at ASC
			LIMIT 1
		`).get() as Record<string, unknown> | undefined;

		if (!row) return null;

		const updated = rawDb.prepare(`
			UPDATE jobs
			SET status='running', claimed_by=?, claimed_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP
			WHERE id=? AND status='queued'
		`).run(workerId, row.id);

		if (updated.changes !== 1) return null;
		return row;
	})();

	if (!job) return new Response(null, { status: 204 });
	return json({ job });
};
