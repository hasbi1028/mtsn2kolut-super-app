import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, apiPost, apiDelete } from '$lib/server/api';

interface GoEmployee {
	id: string; nip: string; nama: string; unit_kerja: string;
	is_active: boolean; created_at: string;
	active_status: string; active_run_type: string;
	last_status: string; last_run_type: string;
}

export const GET: RequestHandler = async () => {
	const items = await apiGet<GoEmployee[]>('/api/employees?with_status=1');
	return json({ items });
};

export const POST: RequestHandler = async ({ request }) => {
	const body = await request.json() as Record<string, unknown>;
	const { nip, nama, unit_kerja = '', pusaka_username, pusaka_password } = body;

	if (!nip || !nama || !pusaka_username || !pusaka_password) {
		return json({ error: 'nip, nama, pusaka_username, pusaka_password wajib diisi' }, { status: 400 });
	}

	try {
		await apiPost('/api/employees', { nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active: true });
		return json({ ok: true }, { status: 201 });
	} catch {
		return json({ error: 'Gagal menambah pegawai (mungkin NIP duplikat)' }, { status: 409 });
	}
};

export const DELETE: RequestHandler = async ({ request }) => {
	const { id } = await request.json() as { id?: string };
	if (!id) return json({ error: 'id wajib diisi' }, { status: 400 });

	try {
		await apiDelete(`/api/employees/${id}`);
		return json({ ok: true });
	} catch {
		return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
	}
};
