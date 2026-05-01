import { redirect, error } from '@sveltejs/kit';
import type { Handle, HandleFetch, RequestEvent } from '@sveltejs/kit';
import { dev } from '$app/environment';
import { env } from '$env/dynamic/private';
import { apiRefreshWithFetch } from '$lib/server/api';
import type { TokenPair } from '$lib/server/api';
import { hasRefreshToken, isAccessTokenValid, getUserFromToken } from '$lib/server/auth';
import { hasAnyRole, isAdminOnlyPath, isKesiswaanPath, isPublicPath, isStaffOperationPath } from '$lib/server/route-access';

const API_BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

function coreApiPathPrefix(basePath: string): string {
	const normalizedBasePath = basePath === '/' ? '' : basePath.replace(/\/$/, '');
	return `${normalizedBasePath}/api`;
}

function isCoreApiRequest(url: string): boolean {
	try {
		const requestUrl = new URL(url);
		const apiUrl = new URL(API_BASE);
		const apiPath = coreApiPathPrefix(apiUrl.pathname);
		return requestUrl.origin === apiUrl.origin
			&& (requestUrl.pathname === apiPath || requestUrl.pathname.startsWith(`${apiPath}/`));
	} catch {
		return false;
	}
}

function hasBearerAuthorization(request: Request): boolean {
	return /^Bearer\s+\S+/i.test(request.headers.get('Authorization') ?? '');
}

function setAuthCookies(event: RequestEvent, tokens: TokenPair) {
	event.cookies.set('access_token', tokens.access_token, {
		path: '/', httpOnly: true, sameSite: 'lax', maxAge: 3600, secure: !dev
	});
	event.cookies.set('refresh_token', tokens.refresh_token, {
		path: '/', httpOnly: true, sameSite: 'lax', maxAge: 7 * 24 * 3600, secure: !dev
	});
	event.locals.accessToken = tokens.access_token;
	event.locals.user = getUserFromToken(tokens.access_token) ?? undefined;
}

function clearAuthCookies(event: RequestEvent) {
	event.cookies.delete('access_token', { path: '/' });
	event.cookies.delete('refresh_token', { path: '/' });
	event.locals.accessToken = undefined;
	event.locals.user = undefined;
}

function requestWithAccessToken(request: Request, accessToken: string): Request {
	const headers = new Headers(request.headers);
	headers.set('Authorization', `Bearer ${accessToken}`);
	return new Request(request, { headers });
}

async function refreshAuthSession(
	event: RequestEvent,
	refreshToken: string,
	fetcher: typeof fetch
): Promise<TokenPair> {
	if (!event.locals.authRefreshPromise) {
		event.locals.authRefreshPromise = apiRefreshWithFetch(fetcher, refreshToken, {
			userAgent: event.request.headers.get('user-agent') ?? '',
			ipAddress: event.getClientAddress(),
		}).then((tokens) => {
			setAuthCookies(event, tokens);
			return tokens;
		}).catch((refreshError: unknown) => {
			clearAuthCookies(event);
			throw refreshError;
		}).finally(() => {
			event.locals.authRefreshPromise = undefined;
		});
	}
	return event.locals.authRefreshPromise;
}

export const handle: Handle = async ({ event, resolve }) => {
	let access = event.cookies.get('access_token');
	const refresh = event.cookies.get('refresh_token');

	if (!isAccessTokenValid(access) && hasRefreshToken(refresh)) {
		try {
			const tokens = await refreshAuthSession(event, refresh!, event.fetch);
			access = tokens.access_token;
		} catch {
			access = undefined;
		}
	}
	if (!isAccessTokenValid(access)) access = undefined;

	event.locals.user = getUserFromToken(access) ?? undefined;
	event.locals.accessToken = access;

	const isPublic = isPublicPath(event.url.pathname);
	if (!isPublic && !event.locals.user) {
		const from = encodeURIComponent(event.url.pathname + event.url.search);
		throw redirect(302, `/login?from=${from}`);
	}

	// Role gate: check if user has admin role in roles array
	if (event.locals.user) {
		const isAdmin = hasAnyRole(event.locals.user, ['admin']);
		
		if (!isAdmin) {
			const isAdminPath = isAdminOnlyPath(event.url.pathname);
			if (isAdminPath) {
				if (event.url.pathname.startsWith('/api/')) {
					throw error(403, 'forbidden: admin role required');
				}
				throw redirect(302, '/');
			}
		}

		if (isStaffOperationPath(event.url.pathname)) {
			const allowed = isAdmin || hasAnyRole(event.locals.user, ['staf']);
			if (!allowed) {
				if (event.url.pathname.startsWith('/api/')) {
					throw error(403, 'forbidden: staf role required');
				}
				throw redirect(302, '/');
			}
		}

		if (isKesiswaanPath(event.url.pathname)) {
			const allowed = isAdmin || hasAnyRole(event.locals.user, ['kesiswaan', 'guru']);
			if (!allowed) {
				if (event.url.pathname.startsWith('/api/')) {
					throw error(403, 'forbidden: kesiswaan role required');
				}
				throw redirect(302, '/');
			}
		}
	}

	return resolve(event);
};

export const handleFetch: HandleFetch = async ({ event, request, fetch }) => {
	const isRefreshRequest = request.url.includes('/api/auth/refresh');
	const refresh = event.cookies.get('refresh_token');
	const canAttemptRefresh = isCoreApiRequest(request.url)
		&& hasBearerAuthorization(request)
		&& !isRefreshRequest
		&& !!refresh
		&& hasRefreshToken(refresh);
	const retrySource = canAttemptRefresh && request.method !== 'GET' && request.method !== 'HEAD' ? request.clone() : request;
	const response = await fetch(request);

	if (response.status !== 401) return response;

	// Don't intercept the refresh endpoint itself to prevent infinite loops
	if (!canAttemptRefresh) return response;

	try {
		const latestAccessToken = event.locals.accessToken;
		const requestAuthorization = request.headers.get('Authorization');
		if (latestAccessToken && isAccessTokenValid(latestAccessToken) && requestAuthorization !== `Bearer ${latestAccessToken}`) {
			return fetch(requestWithAccessToken(retrySource, latestAccessToken));
		}

		const tokens = await refreshAuthSession(event, refresh!, fetch);
		return fetch(requestWithAccessToken(retrySource, tokens.access_token));
	} catch {
		// refreshAuthSession owns cookie cleanup; keep the original 401 for the caller.
	}

	return response;
};
