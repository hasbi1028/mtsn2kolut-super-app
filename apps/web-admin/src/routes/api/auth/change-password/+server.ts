import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, ApiError, handleRouteError } from '$lib/server/api';

interface Body { current_password: string; new_password: string }

export const POST = async (event: RequestEvent) => {
	if (!event.locals.user) return json({ error: 'Unauthorized' }, { status: 401 });

	try {
		const { current_password, new_password } = await event.request.json() as Body;

		if (!current_password || !new_password)
			return json({ error: 'Field tidak boleh kosong' }, { status: 400 });
		if (new_password.length < 8)
			return json({ error: 'Password baru minimal 8 karakter' }, { status: 400 });

		await proxy(event).post('/api/auth/change-password', { username: event.locals.user!.username, old_password: current_password, new_password });
		return json({ ok: true });
	} catch (e) {
		if (e instanceof ApiError && e.status === 401)
			return json({ error: 'Password saat ini salah' }, { status: 401 });
		if (e instanceof ApiError && e.status === 403)
			return json({ error: 'Akun sedang dinonaktifkan' }, { status: 403 });
		return handleRouteError(e, 'auth/change-password');
	}
};
