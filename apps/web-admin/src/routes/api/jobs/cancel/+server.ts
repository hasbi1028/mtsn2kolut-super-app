import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPost, handleRouteError } from '$lib/server/api';

export const POST: RequestHandler = async ({ request }) => {
	try {
		const { employee_id } = await request.json() as { employee_id?: string };
		if (!employee_id) return json({ error: 'employee_id wajib' }, { status: 400 });
		const result = await apiPost<{ ok: boolean; cancelled: number }>('/api/jobs/cancel', { employee_id });
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'jobs/cancel');
	}
};
