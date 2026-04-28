import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, ApiError, handleRouteError } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const { id } = event.params;
		const { pusaka_username, pusaka_password } = await event.request.json() as {
			pusaka_username?: string;
			pusaka_password?: string;
		};

		if (!pusaka_username) return json({ error: 'pusaka_username wajib diisi' }, { status: 400 });

		const result = await proxy(event).post(`/api/employees/${id}/update-pusaka`, {
			pusaka_username,
			pusaka_password: pusaka_password ?? '',
		});
		return json(result);
	} catch (e) {
		if (e instanceof ApiError && e.status === 404)
			return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });
		return handleRouteError(e, 'employees/:id PUT');
	}
};
