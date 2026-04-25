import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { rawDb } from '$lib/server/db';

export const POST: RequestHandler = async ({ request }) => {
	const { employee_id } = await request.json() as { employee_id?: string };
	if (!employee_id) return json({ error: 'employee_id wajib' }, { status: 400 });

	const result = rawDb.prepare(
		`UPDATE jobs SET status='failed', error_message='Dibatalkan manual', next_retry_at=NULL, updated_at=CURRENT_TIMESTAMP
		 WHERE employee_id=? AND status IN ('queued','running')`
	).run(employee_id);

	return json({ ok: true, cancelled: result.changes });
};
