import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, ApiError, handleRouteError } from '$lib/server/api';

interface GoEmployee {
	id: string; nip: string; nama: string; unit_kerja: string;
	pusaka_username: string; is_active: boolean; created_at: string;
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
		const body = await event.request.json() as Record<string, unknown>;
		const { nip, nama, unit_kerja = '', pusaka_username, pusaka_password } = body;

		if (!nip || !nama || !pusaka_username || !pusaka_password)
			return json({ error: 'nip, nama, pusaka_username, pusaka_password wajib diisi' }, { status: 400 });

		await proxy(event).post('/api/employees', { nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active: true });
		return json({ ok: true }, { status: 201 });
	} catch (e) {
		if (e instanceof ApiError && e.status === 409)
			return json({ error: 'Gagal menambah pegawai (NIP sudah terdaftar)' }, { status: 409 });
		return handleRouteError(e, 'employees POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const { id } = await event.request.json() as { id?: string };
		if (!id) return json({ error: 'id wajib diisi' }, { status: 400 });

		await proxy(event).del(`/api/employees/${id}`);
		return json({ ok: true });
	} catch (e) {
		if (e instanceof ApiError && e.status === 404)
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		return handleRouteError(e, 'employees DELETE');
	}
};
