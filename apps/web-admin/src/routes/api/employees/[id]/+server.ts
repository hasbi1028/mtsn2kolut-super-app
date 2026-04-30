import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, ApiError, handleRouteError } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const { id } = event.params;
		const body = await event.request.json() as Record<string, unknown>;
		const { nip, nama, unit_kerja = '', employment_type, is_active } = body;

		if (!nip || !nama || !employment_type) {
			return json({ error: 'nip, nama, dan employment_type wajib diisi' }, { status: 400 });
		}
		if (typeof is_active !== 'boolean') {
			return json({ error: 'is_active wajib boolean' }, { status: 400 });
		}

		const result = await proxy(event).put(`/api/employees/${id}`, {
			nip,
			nama,
			unit_kerja,
			employment_type,
			is_active,
		});
		return json(result);
	} catch (e) {
		if (e instanceof ApiError && e.status === 404) {
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		}
		if (e instanceof ApiError && e.status === 400) {
			return json({ error: e.message }, { status: 400 });
		}
		return handleRouteError(e, 'employees/:id PUT');
	}
};
