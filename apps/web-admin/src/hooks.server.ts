import { redirect, error } from '@sveltejs/kit';
import type { Handle, HandleFetch, HandleServerError, RequestEvent } from '@sveltejs/kit';
import { dev } from '$app/environment';
import { env } from '$env/dynamic/private';
import { ApiError, AuthValidationUnavailableError, apiRefreshWithFetch, getVerifiedUserFromAccessToken } from '$lib/server/api';
import type { TokenPair } from '$lib/server/api';
import { hasRefreshToken, isAccessTokenValid, getUserFromToken } from '$lib/server/auth';
import { canAccessProtectedRoute, isMustChangePasswordAllowedPath, isPublicPath } from '$lib/server/route-access';

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
	event.locals.user = undefined;
	event.locals.authUserPromise = undefined;
}

function clearAuthCookies(event: RequestEvent) {
	event.cookies.delete('access_token', { path: '/' });
	event.cookies.delete('refresh_token', { path: '/' });
	event.locals.accessToken = undefined;
	event.locals.user = undefined;
	event.locals.authUserPromise = undefined;
}

function isExplicitAuthRejection(error: unknown): boolean {
	return error instanceof ApiError && (error.status === 401 || error.status === 403);
}

async function verifiedUserForAccessToken(event: RequestEvent, accessToken: string) {
	if (!event.locals.authUserPromise) {
		event.locals.authUserPromise = getVerifiedUserFromAccessToken(event.fetch, accessToken, getUserFromToken);
	}
	return event.locals.authUserPromise;
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
			if (isExplicitAuthRejection(refreshError)) clearAuthCookies(event);
			throw refreshError;
		}).finally(() => {
			event.locals.authRefreshPromise = undefined;
		});
	}
	return event.locals.authRefreshPromise;
}

function staticCacheControlForPath(pathname: string): string | null {
	if (pathname.startsWith('/_app/immutable/')) return 'public, max-age=31536000, immutable';
	if (pathname === '/favicon.svg' || pathname === '/manifest.webmanifest') return 'public, max-age=86400, stale-while-revalidate=604800';
	if (pathname.startsWith('/releases/mobile/') && !pathname.endsWith('.apk')) return 'public, max-age=300, stale-while-revalidate=3600';
	return null;
}

function isStateChangingMethod(method: string): boolean {
	return !['GET', 'HEAD', 'OPTIONS'].includes(method.toUpperCase());
}

function originMatchesRequest(event: RequestEvent): boolean {
	const origin = event.request.headers.get('origin');
	if (origin) return origin === event.url.origin;
	const referer = event.request.headers.get('referer');
	if (!referer) return false;
	try {
		return new URL(referer).origin === event.url.origin;
	} catch {
		return false;
	}
}

function isProtectedBffMutation(event: RequestEvent, isPublic: boolean): boolean {
	return event.url.pathname.startsWith('/api/') && !isPublic && isStateChangingMethod(event.request.method);
}

export const handle: Handle = async ({ event, resolve }) => {
	const isPublic = isPublicPath(event.url.pathname);
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

	event.locals.user = undefined;
	event.locals.accessToken = access;
	if (access) {
		let validationUnavailable = false;
		try {
			event.locals.user = await verifiedUserForAccessToken(event, access);
		} catch (validationError) {
			if (validationError instanceof AuthValidationUnavailableError) {
				if (!isPublic) {
					throw error(503, 'Layanan validasi sesi sedang bermasalah. Silakan coba beberapa saat lagi.');
				}
				validationUnavailable = true;
				event.locals.user = undefined;
			} else {
				throw validationError;
			}
		}
		if (!event.locals.user && !validationUnavailable && hasRefreshToken(refresh)) {
			try {
				const tokens = await refreshAuthSession(event, refresh!, event.fetch);
				access = tokens.access_token;
				event.locals.accessToken = access;
				event.locals.user = await verifiedUserForAccessToken(event, access);
			} catch (validationError) {
				if (validationError instanceof AuthValidationUnavailableError) {
					if (!isPublic) {
						throw error(503, 'Layanan validasi sesi sedang bermasalah. Silakan coba beberapa saat lagi.');
					}
					validationUnavailable = true;
					event.locals.user = undefined;
				} else if (isExplicitAuthRejection(validationError)) {
					access = undefined;
				} else {
					if (!isPublic) {
						throw error(503, 'Layanan sesi sedang bermasalah. Silakan coba beberapa saat lagi.');
					}
					validationUnavailable = true;
				}
			}
		}
		if (!event.locals.user && !validationUnavailable) {
			clearAuthCookies(event);
			access = undefined;
		}
	}

	if (!isPublic && !event.locals.user) {
		if (event.url.pathname.startsWith('/api/')) {
			return new Response(JSON.stringify({ error: 'unauthorized' }), {
				status: 401,
				headers: { 'Content-Type': 'application/json' }
			});
		}
		const from = encodeURIComponent(event.url.pathname + event.url.search);
		throw redirect(302, `/login?from=${from}`);
	}

	if (event.locals.user?.must_change_password && !isMustChangePasswordAllowedPath(event.url.pathname, event.request.method)) {
		if (event.url.pathname.startsWith('/api/')) {
			throw error(403, 'password change required');
		}
		throw redirect(302, '/settings/account');
	}

	if (isProtectedBffMutation(event, isPublic) && !originMatchesRequest(event)) {
		throw error(403, 'csrf validation failed');
	}

	// Permission-aware gate: prefer dynamic RBAC permissions, keep legacy role fallback during migration.
	if (event.locals.user && !canAccessProtectedRoute(event.locals.user, event.url.pathname, event.request.method)) {
		if (event.url.pathname.startsWith('/api/')) {
			throw error(403, 'forbidden: insufficient permission');
		}
		throw redirect(302, '/');
	}

	const cacheControl = staticCacheControlForPath(event.url.pathname);
	const response = await resolve(event);
	if (cacheControl && !response.headers.has('cache-control')) {
		response.headers.set('cache-control', cacheControl);
	}
	return response;
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

export const handleError: HandleServerError = ({ error, event, status, message }) => {
	const reference = `WEB-${Date.now().toString(36).toUpperCase()}`;
	console.error(`[web-admin ${reference}] ${event.request.method} ${event.url.pathname} -> ${status}`, error);

	if (event.url.pathname.startsWith('/api/')) {
		return {
			message: status >= 500
				? `Layanan sistem sedang bermasalah. Kode referensi: ${reference}`
				: message
		};
	}

	return {
		message: status >= 500
			? `Halaman sedang bermasalah. Data tetap aman. Kode referensi: ${reference}`
			: message
	};
};
