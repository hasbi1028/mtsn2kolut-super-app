import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPost, ApiError, handleRouteError } from '$lib/server/api';

// PUT /api/employees/:id — update pusaka credentials via dedicated backend endpoint
export const PUT: RequestHandler = async ({ params, request }) => {
	try {
		const { id } = params;
		const { pusaka_username, pusaka_password } = await request.json() as {
			pusaka_username?: string;
			pusaka_password?: string;
		};

		if (!pusaka_username) return json({ error: 'pusaka_username wajib diisi' }, { status: 400 });

		const result = await apiPost(`/api/employees/${id}/update-pusaka`, {
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
