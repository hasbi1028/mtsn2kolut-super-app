import { dev } from '$app/environment';
import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit';
import { apiRefresh } from '$lib/server/api';
import { hasRefreshToken, isAccessTokenValid } from '$lib/server/auth';

const PUBLIC_PREFIXES = ['/login', '/api/auth/logout'];

export const handle: Handle = async ({ event, resolve }) => {
	const accessToken = event.cookies.get('access_token');
	const refreshToken = event.cookies.get('refresh_token');

	if (isAccessTokenValid(accessToken)) {
		event.locals.user = { id: 'admin' };
	} else if (hasRefreshToken(refreshToken)) {
		try {
			const pair = await apiRefresh(refreshToken!);
			event.cookies.set('access_token', pair.access_token, {
				path: '/', httpOnly: true, sameSite: 'lax',
				secure: !dev,
				maxAge: 60 * 60,
			});
			event.cookies.set('refresh_token', pair.refresh_token, {
				path: '/', httpOnly: true, sameSite: 'lax',
				secure: !dev,
				maxAge: 7 * 24 * 60 * 60,
			});
			event.locals.user = { id: 'admin' };
		} catch {
			event.cookies.delete('access_token', { path: '/' });
			event.cookies.delete('refresh_token', { path: '/' });
		}
	}

	const isPublic = PUBLIC_PREFIXES.some((p) => event.url.pathname.startsWith(p));
	if (!isPublic && !event.locals.user) {
		const from = encodeURIComponent(event.url.pathname + event.url.search);
		throw redirect(302, `/login?from=${from}`);
	}

	return resolve(event);
};
