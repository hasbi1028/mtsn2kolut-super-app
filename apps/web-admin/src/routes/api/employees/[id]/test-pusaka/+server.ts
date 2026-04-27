import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, ApiError, handleRouteError } from '$lib/server/api';

// POST /api/employees/:id/test-pusaka — checks if credentials are configured
export const POST: RequestHandler = async ({ params }) => {
	try {
		const { id } = params;
		const data = await apiGet<{ configured: boolean; pusaka_username: string }>(`/api/employees/${id}/pusaka-status`);

		if (!data.configured) {
			return json({ message: `Kredensial belum dikonfigurasi untuk pegawai ini.` }, { status: 200 });
		}
		return json({ message: `Kredensial Pusaka tersimpan untuk: ${data.pusaka_username}` }, { status: 200 });
	} catch (e) {
		if (e instanceof ApiError && e.status === 404)
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		return handleRouteError(e, 'employees/:id/test-pusaka');
	}
};
