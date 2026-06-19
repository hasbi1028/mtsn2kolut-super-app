import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const { employee_id, run_type = 'morning', max_attempts = 3 } =
			await readRequestJson<{ employee_id?: string; run_type?: string; max_attempts?: number }>(event.request);
		if (!employee_id) return json({ error: 'employee_id wajib' }, { status: 400 });
		await proxy(event).post('/api/pusaka/jobs', { employee_id, run_type, max_attempts });
		return json({ inserted: 1, skipped: 0 }, { status: 201 });
	} catch (e) {
		if (e instanceof ApiError && e.status === 409)
			return json({ inserted: 0, skipped: 1, reason: 'job queued/running already exists' });
		return handleRouteError(e, 'pusaka/jobs/run-now');
	}
};
