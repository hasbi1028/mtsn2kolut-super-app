import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { ApiError, apiPath, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

interface GoEmployee {
	id: string; nip: string; nama: string; unit_kerja: string;
	employment_type: string;
	pusaka_username: string; pusaka_eligible: boolean; has_pusaka_account: boolean;
	is_active: boolean; created_at: string;
	active_status: string; active_run_type: string;
	last_status: string; last_run_type: string;
	has_checkin_schedule: boolean; has_checkout_schedule: boolean;
}

export const GET = async (event: RequestEvent) => {
	try {
		const items = await proxy(event).get<GoEmployee[]>('/api/employees?with_status=1');
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'employees GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const { nip, nama, unit_kerja = '', employment_type = 'lainnya', pusaka_username, pusaka_password } = body;

		if (!nip || !nama)
			return json({ error: 'nip dan nama wajib diisi' }, { status: 400 });
		if ((pusaka_username && !pusaka_password) || (!pusaka_username && pusaka_password))
			return json({ error: 'username dan password PUSAKA harus diisi berpasangan' }, { status: 400 });
		if (!['pns', 'pppk', 'honorer', 'lainnya'].includes(String(employment_type)))
			return json({ error: 'employment_type tidak valid' }, { status: 400 });

		await proxy(event).post('/api/employees', { nip, nama, unit_kerja, employment_type, pusaka_username, pusaka_password, is_active: true });
		return json({ ok: true }, { status: 201 });
	} catch (e) {
		if (e instanceof ApiError && e.status === 409)
			return json({ error: 'Gagal menambah pegawai (NIP sudah terdaftar)' }, { status: 409 });
		if (e instanceof ApiError && e.status === 400)
			return json({ error: e.message }, { status: 400 });
		return handleRouteError(e, 'employees POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const { id } = await readRequestJson<{ id?: string }>(event.request);
		if (!id) return json({ error: 'id wajib diisi' }, { status: 400 });

		await proxy(event).del(apiPath`/api/employees/${id}`);
		return json({ ok: true });
	} catch (e) {
		if (e instanceof ApiError && e.status === 404)
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		return handleRouteError(e, 'employees DELETE');
	}
};
