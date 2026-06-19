import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { apiPublicPostWithFetch } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	const refreshToken = event.cookies.get('refresh_token') ?? '';

	if (refreshToken) {
		try {
			await apiPublicPostWithFetch(event.fetch, '/api/auth/logout', { refresh_token: refreshToken });
		} catch {
			// Best effort revoke — cookies are still cleared locally below.
		}
	}

	event.cookies.delete('access_token', { path: '/' });
	event.cookies.delete('refresh_token', { path: '/' });
	return json({ ok: true });
};
