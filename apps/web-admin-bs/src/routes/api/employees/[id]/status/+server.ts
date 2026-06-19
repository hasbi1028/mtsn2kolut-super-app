import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PATCH = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<{ is_active?: boolean }>(event.request);
		if (typeof body.is_active !== 'boolean') {
			return json({ error: 'is_active wajib boolean' }, { status: 400 });
		}
		const data = await proxy(event).patch(apiPath`/api/employees/${id}/status`, body);
		return json(data);
	} catch (e) {
		if (e instanceof ApiError && e.status === 404) {
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		}
		return handleRouteError(e, 'employees/:id/status PATCH');
	}
};
