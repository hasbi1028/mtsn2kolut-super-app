import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPost, ApiError, handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async ({ request }) => {
	try {
		const { employee_id, run_type = 'morning', max_attempts = 3 } =
			await request.json() as { employee_id?: string; run_type?: string; max_attempts?: number };

		if (!employee_id) return json({ error: 'employee_id wajib' }, { status: 400 });

		await apiPost('/api/jobs', { employee_id, run_type, max_attempts });
		return json({ inserted: 1, skipped: 0 }, { status: 201 });
	} catch (e) {
		if (e instanceof ApiError && e.status === 409)
			return json({ inserted: 0, skipped: 1, reason: 'job queued/running already exists' });
		return handleRouteError(e, 'jobs/run-now');
	}
};
