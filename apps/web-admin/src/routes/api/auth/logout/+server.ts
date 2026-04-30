import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { apiPost } from '$lib/server/api';

export const POST: RequestHandler = async ({ cookies }) => {
	const refreshToken = cookies.get('refresh_token') ?? '';

	if (refreshToken) {
		try {
			await apiPost('/api/auth/logout', { refresh_token: refreshToken });
		} catch {
			// Best effort revoke — cookies are still cleared locally below.
		}
	}

	cookies.delete('access_token', { path: '/' });
	cookies.delete('refresh_token', { path: '/' });
	return json({ ok: true });
};
