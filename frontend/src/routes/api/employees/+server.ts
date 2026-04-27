import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, apiPost, apiDelete, ApiError, handleRouteError } from '$lib/server/api';

interface GoEmployee {
	id: string; nip: string; nama: string; unit_kerja: string;
	is_active: boolean; created_at: string;
	active_status: string; active_run_type: string;
	last_status: string; last_run_type: string;
}

export const GET: RequestHandler = async () => {
	try {
		const items = await apiGet<GoEmployee[]>('/api/employees?with_status=1');
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'employees GET');
	}
};

export const POST: RequestHandler = async ({ request }) => {
	try {
		const body = await request.json() as Record<string, unknown>;
		const { nip, nama, unit_kerja = '', pusaka_username, pusaka_password } = body;

		if (!nip || !nama || !pusaka_username || !pusaka_password)
			return json({ error: 'nip, nama, pusaka_username, pusaka_password wajib diisi' }, { status: 400 });

		await apiPost('/api/employees', { nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active: true });
		return json({ ok: true }, { status: 201 });
	} catch (e) {
		if (e instanceof ApiError && e.status === 409)
			return json({ error: 'Gagal menambah pegawai (NIP sudah terdaftar)' }, { status: 409 });
		return handleRouteError(e, 'employees POST');
	}
};

export const DELETE: RequestHandler = async ({ request }) => {
	try {
		const { id } = await request.json() as { id?: string };
		if (!id) return json({ error: 'id wajib diisi' }, { status: 400 });

		await apiDelete(`/api/employees/${id}`);
		return json({ ok: true });
	} catch (e) {
		if (e instanceof ApiError && e.status === 404)
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		return handleRouteError(e, 'employees DELETE');
	}
};
