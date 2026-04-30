import { redirect, error } from '@sveltejs/kit';
import type { Handle, HandleFetch } from '@sveltejs/kit';
import { dev } from '$app/environment';
import { apiRefresh } from '$lib/server/api';
import { hasRefreshToken, isAccessTokenValid, getUserFromToken } from '$lib/server/auth';

const PUBLIC_EXACT_PATHS = new Set([
	'/',
	'/login',
	'/ppdb',
	'/profil',
	'/berita',
	'/pengumuman',
	'/kontak',
	'/api/auth/logout',
	'/api/public/register-student',
	'/api/public/site/posts',
	'/api/public/site/announcements',
]);

const PUBLIC_PREFIXES = [
	'/berita/',
	'/pengumuman/',
	'/api/public/site/pages/',
	'/api/public/site/posts/',
	'/api/public/site/announcements/',
];

// Admin-only path prefixes (UI pages + API proxy routes).
// Guru users get redirected (UI) or 403 (API).
const ADMIN_PREFIXES = [
	'/employees',
	'/academic',
	'/pusaka',
	'/settings',
	'/website',
	'/api/employees',
	'/api/users',
	'/api/pusaka',
	'/api/website',
];

function isPublicPath(pathname: string) {
	if (PUBLIC_EXACT_PATHS.has(pathname)) return true;
	return PUBLIC_PREFIXES.some((p) => pathname.startsWith(p));
}

export const handle: Handle = async ({ event, resolve }) => {
	let access = event.cookies.get('access_token');
	const refresh = event.cookies.get('refresh_token');

	if (!isAccessTokenValid(access) && hasRefreshToken(refresh)) {
		try {
			const tokens = await apiRefresh(refresh!, {
				userAgent: event.request.headers.get('user-agent') ?? '',
				ipAddress: event.getClientAddress(),
			});
			event.cookies.set('access_token', tokens.access_token, {
				path: '/', httpOnly: true, sameSite: 'lax', maxAge: 3600, secure: !dev
			});
			event.cookies.set('refresh_token', tokens.refresh_token, {
				path: '/', httpOnly: true, sameSite: 'lax', maxAge: 7 * 24 * 3600, secure: !dev
			});
			access = tokens.access_token;
		} catch {
			event.cookies.delete('access_token', { path: '/' });
			event.cookies.delete('refresh_token', { path: '/' });
		}
	}

	event.locals.user = getUserFromToken(access) as any;
	event.locals.accessToken = access;

	const isPublic = isPublicPath(event.url.pathname);
	if (!isPublic && !event.locals.user) {
		const from = encodeURIComponent(event.url.pathname + event.url.search);
		throw redirect(302, `/login?from=${from}`);
	}

	// Role gate: check if user has admin role in roles array
	if (event.locals.user) {
		const roles = event.locals.user.roles || [];
		const isAdmin = roles.includes('admin');
		
		if (!isAdmin) {
			const isAdminPath = ADMIN_PREFIXES.some((p) => event.url.pathname.startsWith(p));
			if (isAdminPath) {
				if (event.url.pathname.startsWith('/api/')) {
					throw error(403, 'forbidden: admin role required');
				}
				throw redirect(302, '/');
			}
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
		const tokens = await apiRefresh(refresh, {
			userAgent: event.request.headers.get('user-agent') ?? '',
			ipAddress: event.getClientAddress(),
		});
		event.cookies.set('access_token', tokens.access_token, {
			path: '/', httpOnly: true, sameSite: 'lax', maxAge: 3600, secure: !dev
		});
		event.cookies.set('refresh_token', tokens.refresh_token, {
			path: '/', httpOnly: true, sameSite: 'lax', maxAge: 7 * 24 * 3600, secure: !dev
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
