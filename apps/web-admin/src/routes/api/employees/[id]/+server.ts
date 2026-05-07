import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const {
			nip = '',
			nama,
			unit_kerja = '',
			employment_type,
			tanggal_lahir = '',
			jenis_kelamin = '',
			tempat_lahir = '',
			is_active
		} = body;

		if (!nama || !employment_type) {
			return json({ error: 'nama dan employment_type wajib diisi' }, { status: 400 });
		}
		if (typeof is_active !== 'boolean') {
			return json({ error: 'is_active wajib boolean' }, { status: 400 });
		}
		if (jenis_kelamin && !['L', 'P'].includes(String(jenis_kelamin))) {
			return json({ error: 'jenis_kelamin tidak valid' }, { status: 400 });
		}

		const result = await proxy(event).put(apiPath`/api/employees/${id}`, {
			nip,
			nama,
			unit_kerja,
			employment_type,
			tanggal_lahir,
			jenis_kelamin,
			tempat_lahir,
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
		if (e instanceof ApiError && e.status === 409) {
			return json({ error: e.message }, { status: 409 });
		}
		return handleRouteError(e, 'employees/:id PUT');
	}
};
