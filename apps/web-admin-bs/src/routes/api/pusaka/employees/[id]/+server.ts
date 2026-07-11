import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const { pusaka_username, pusaka_password } = await readRequestJson<{
			pusaka_username?: string;
			pusaka_password?: string;
		}>(event.request);

		if (!pusaka_username) return json({ error: 'pusaka_username wajib diisi' }, { status: 400 });

		const result = await proxy(event).post(apiPath`/api/pusaka/employees/${id}/update-pusaka`, {
			pusaka_username,
			pusaka_password: pusaka_password ?? '',
		});
		return json(result);
	} catch (e) {
		if (e instanceof ApiError && e.status === 404)
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		if (e instanceof ApiError && e.status === 400)
			return json({ error: e.message }, { status: 400 });
		return handleRouteError(e, 'pusaka/employees/:id PUT');
	}
};

export const PATCH = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const { is_enabled } = await readRequestJson<{ is_enabled?: boolean }>(event.request);
		if (typeof is_enabled !== 'boolean') return json({ error: 'is_enabled wajib boolean' }, { status: 400 });
		const result = await proxy(event).patch(apiPath`/api/pusaka/employees/${id}/account-status`, { is_enabled });
		return json(result);
	} catch (e) {
		if (e instanceof ApiError && e.status === 404)
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		if (e instanceof ApiError && e.status === 400)
			return json({ error: e.message }, { status: 400 });
		return handleRouteError(e, 'pusaka/employees/:id PATCH');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const result = await proxy(event).del(apiPath`/api/pusaka/employees/${id}/account`);
		return json(result);
	} catch (e) {
		if (e instanceof ApiError && e.status === 404)
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		if (e instanceof ApiError && e.status === 400)
			return json({ error: e.message }, { status: 400 });
		return handleRouteError(e, 'pusaka/employees/:id DELETE');
	}
};
