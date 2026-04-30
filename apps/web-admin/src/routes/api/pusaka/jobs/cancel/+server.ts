import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const { employee_id } = await event.request.json() as { employee_id?: string };
		if (!employee_id) return json({ error: 'employee_id wajib' }, { status: 400 });
		const result = await proxy(event).post<{ ok: boolean; cancelled: number }>('/api/pusaka/jobs/cancel', { employee_id });
		return json(result);
	} catch (e) {
		return handleRouteError(e, 'pusaka/jobs/cancel');
	}
};
