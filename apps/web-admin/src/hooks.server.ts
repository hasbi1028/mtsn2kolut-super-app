import { redirect, error } from '@sveltejs/kit';
import type { Handle, HandleFetch } from '@sveltejs/kit';
import { apiRefresh } from '$lib/server/api';
import { hasRefreshToken, isAccessTokenValid, getUserFromToken } from '$lib/server/auth';

const PUBLIC_PREFIXES = ['/login', '/api/auth/logout'];

// Admin-only path prefixes (UI pages + API proxy routes).
// Guru users get redirected (UI) or 403 (API).
const ADMIN_PREFIXES = [
	'/employees',
	'/academic',
	'/pusaka',
	'/settings',
	'/api/employees',
	'/api/users',
	'/api/jobs',
	'/api/attendance',
	'/api/schedules',
	'/api/settings',
	'/api/scheduler',
];

export const handle: Handle = async ({ event, resolve }) => {
	let access = event.cookies.get('access_token');
	const refresh = event.cookies.get('refresh_token');

	if (!isAccessTokenValid(access) && hasRefreshToken(refresh)) {
		try {
			const tokens = await apiRefresh(refresh!);
			event.cookies.set('access_token', tokens.access_token, {
				path: '/', httpOnly: true, sameSite: 'lax', maxAge: 3600, secure: false
			});
			event.cookies.set('refresh_token', tokens.refresh_token, {
				path: '/', httpOnly: true, sameSite: 'lax', maxAge: 7 * 24 * 3600, secure: false
			});
			access = tokens.access_token;
		} catch {
			event.cookies.delete('access_token', { path: '/' });
			event.cookies.delete('refresh_token', { path: '/' });
		}
	}

	event.locals.user = getUserFromToken(access) as any;
	event.locals.accessToken = access;

	const isPublic = PUBLIC_PREFIXES.some((p) => event.url.pathname.startsWith(p));
	if (!isPublic && !event.locals.user) {
		const from = encodeURIComponent(event.url.pathname + event.url.search);
		throw redirect(302, `/login?from=${from}`);
	}

	// Role gate: non-admin users cannot access admin routes
	if (event.locals.user && event.locals.user.role !== 'admin') {
		const isAdminPath = ADMIN_PREFIXES.some((p) => event.url.pathname.startsWith(p));
		if (isAdminPath) {
			if (event.url.pathname.startsWith('/api/')) {
				throw error(403, 'forbidden: admin role required');
			}
			throw redirect(302, '/');
		}
	}

	return resolve(event);
};

export const handleFetch: HandleFetch = async ({ event, request, fetch }) => {
	const response = await fetch(request);

	if (response.status !== 401) return response;

	// Don't intercept the refresh endpoint itself to prevent infinite loops
	if (request.url.includes('/api/auth/refresh')) return response;

	const refresh = event.cookies.get('refresh_token');
	if (!refresh || !hasRefreshToken(refresh)) return response;

	try {
		const tokens = await apiRefresh(refresh);
		event.cookies.set('access_token', tokens.access_token, {
			path: '/', httpOnly: true, sameSite: 'lax', maxAge: 3600, secure: false
		});
		event.cookies.set('refresh_token', tokens.refresh_token, {
			path: '/', httpOnly: true, sameSite: 'lax', maxAge: 7 * 24 * 3600, secure: false
		});

		const newHeaders = new Headers(request.headers);
		newHeaders.set('Authorization', `Bearer ${tokens.access_token}`);

		const retryRequest = new Request(request, { headers: newHeaders });
		return fetch(retryRequest);
	} catch {
		event.cookies.delete('access_token', { path: '/' });
		event.cookies.delete('refresh_token', { path: '/' });
	}

	return response;
};
