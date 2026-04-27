import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPost } from '$lib/server/api';

export const POST: RequestHandler = async ({ request }) => {
	const { employee_id } = await request.json() as { employee_id?: string };
	if (!employee_id) return json({ error: 'employee_id wajib' }, { status: 400 });

	await apiPost('/api/jobs/cancel', { employee_id });
	return json({ ok: true });
};
