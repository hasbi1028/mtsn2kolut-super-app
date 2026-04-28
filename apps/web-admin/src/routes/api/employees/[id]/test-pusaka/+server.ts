import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, ApiError, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const { id } = event.params;
		const data = await proxy(event).get<{ configured: boolean; pusaka_username: string }>(`/api/employees/${id}/pusaka-status`);

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
