import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, ApiError, handleRouteError } from '$lib/server/api';

export const PATCH = async (event: RequestEvent) => {
	try {
		const { id } = event.params;
		const body = await event.request.json() as { is_active?: boolean };
		if (typeof body.is_active !== 'boolean') {
			return json({ error: 'is_active wajib boolean' }, { status: 400 });
		}
		const data = await proxy(event).patch(`/api/employees/${id}/status`, body);
		return json(data);
	} catch (e) {
		if (e instanceof ApiError && e.status === 404) {
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		}
		return handleRouteError(e, 'employees/:id/status PATCH');
	}
};
