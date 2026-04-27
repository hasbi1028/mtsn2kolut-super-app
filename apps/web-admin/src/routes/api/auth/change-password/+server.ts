import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPost, ApiError, handleRouteError } from '$lib/server/api';

interface Body { current_password: string; new_password: string }

export const POST: RequestHandler = async ({ request, locals }) => {
	if (!locals.user) return json({ error: 'Unauthorized' }, { status: 401 });

	try {
		const { current_password, new_password } = await request.json() as Body;

		if (!current_password || !new_password)
			return json({ error: 'Field tidak boleh kosong' }, { status: 400 });
		if (new_password.length < 6)
			return json({ error: 'Password baru minimal 6 karakter' }, { status: 400 });

		await apiPost('/api/auth/change-password', { username: 'admin', old_password: current_password, new_password });
		return json({ ok: true });
	} catch (e) {
		if (e instanceof ApiError && e.status === 401)
			return json({ error: 'Password saat ini salah' }, { status: 401 });
		return handleRouteError(e, 'auth/change-password');
	}
};
