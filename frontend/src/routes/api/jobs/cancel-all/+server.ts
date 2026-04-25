import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { rawDb } from '$lib/server/db';

export const POST: RequestHandler = () => {
	const result = rawDb.prepare(
		`UPDATE jobs SET status='failed', error_message='Dibatalkan manual', next_retry_at=NULL, updated_at=CURRENT_TIMESTAMP
		 WHERE status IN ('queued','running')`
	).run();

	return json({ ok: true, cancelled: result.changes });
};
