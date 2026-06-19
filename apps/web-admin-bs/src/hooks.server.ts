import type { Handle, HandleFetch } from '@sveltejs/kit';
import {
  apiRefreshWithFetch,
  apiValidateAuthWithFetch,
  AuthValidationUnavailableError,
} from '$lib/server/api';
import {
  isAccessTokenValid,
  hasRefreshToken,
  getUserFromToken,
} from '$lib/server/auth';
import type { AuthUser } from '$lib/server/auth';
import {
  isPublicPath,
  canAccessProtectedRoute,
  isMustChangePasswordAllowedPath,
} from '$lib/server/route-access';
import type { TokenPair } from '$lib/server/api';

function isBearerAuth(request: Request): boolean {
  const auth = request.headers.get('Authorization');
  return !!auth && auth.startsWith('Bearer ');
}

function isLocalhostOrigin(urlStr: string): boolean {
  try {
    const u = new URL(urlStr);
    return u.hostname === 'localhost' || u.hostname === '127.0.0.1';
  } catch {
    return false;
  }
}

type RequestEvent = Parameters<Handle>[0]['event'];

const refreshPromises = new WeakMap<object, Promise<TokenPair | null>>();

function setCookies(event: RequestEvent, pair: TokenPair) {
  event.cookies.set('access_token', pair.access_token, {
    path: '/', httpOnly: true, sameSite: 'lax', maxAge: 3600,
  });
  event.cookies.set('refresh_token', pair.refresh_token, {
    path: '/', httpOnly: true, sameSite: 'lax', maxAge: 7 * 24 * 3600,
  });
}

function clearCookies(event: RequestEvent) {
  event.cookies.delete('access_token', { path: '/' });
  event.cookies.delete('refresh_token', { path: '/' });
}

/** Return a 302 redirect Response — avoids importing redirect() from the circular main chunk */
function redirectTo(location: string): Response {
  return new Response(null, { status: 302, headers: { location } });
}

/** Return a JSON error Response — avoids importing error() from the circular main chunk */
function jsonError(status: number, message?: string): Response {
  const body = message ? JSON.stringify({ error: message }) : JSON.stringify({ error: 'error' });
  return new Response(body, { status, headers: { 'content-type': 'application/json' } });
}

async function tryRefresh(event: RequestEvent, fetcher: typeof fetch): Promise<TokenPair | null> {
  const existing = refreshPromises.get(event as object);
  if (existing) return existing;

  const refreshToken = event.cookies.get('refresh_token');
  if (!refreshToken || !hasRefreshToken(refreshToken)) return null;

  const promise = (async () => {
    try {
      const pair = await apiRefreshWithFetch(fetcher, refreshToken, {
        userAgent: event.request.headers.get('user-agent') || '',
        ipAddress: event.getClientAddress?.() || '',
      });
      setCookies(event, pair);
      return pair;
    } catch {
      return null;
    }
  })();

  refreshPromises.set(event as object, promise);
  return promise;
}

export const handleFetch: HandleFetch = async ({ event, request, fetch: routeFetch }) => {
  if (!isBearerAuth(request) || !isLocalhostOrigin(request.url)) {
    return routeFetch(request);
  }

  const requestClone = request.body ? request.clone() : null;
  const response = await routeFetch(request);
  if (response.status !== 401) return response;

  const pair = await tryRefresh(event, routeFetch);
  if (!pair) {
    clearCookies(event);
    return response;
  }

  event.locals.accessToken = pair.access_token;

  const originalBody = requestClone ? await requestClone.text() : undefined;
  const retryHeaders = new Headers(request.headers);
  retryHeaders.set('Authorization', `Bearer ${pair.access_token}`);

  const retryRequest = new Request(request.url, {
    method: request.method,
    headers: retryHeaders,
    body: originalBody,
  });

  return routeFetch(retryRequest);
};

const CSRF_MUTATION_METHODS = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);

export const handle: Handle = async ({ event, resolve }) => {
  const { pathname } = event.url;
  const method = event.request.method;

  if (isPublicPath(pathname)) {
    return resolve(event);
  }

  const accessToken = event.cookies.get('access_token');
  const refreshToken = event.cookies.get('refresh_token');

  if (accessToken && isAccessTokenValid(accessToken)) {
    event.locals.accessToken = accessToken;
    const authUser = getUserFromToken(accessToken);

    if (authUser) {
      try {
        const valid = await apiValidateAuthWithFetch(event.fetch, accessToken);
        if (!valid) {
          clearCookies(event);
          if (pathname.startsWith('/api/')) {
            return jsonError(401, 'unauthorized');
          }
          return redirectTo(`/login?from=${encodeURIComponent(pathname)}`);
        }
      } catch (e) {
        if (e instanceof AuthValidationUnavailableError) {
          return jsonError(503, 'Layanan validasi sesi sedang bermasalah. Silakan coba beberapa saat lagi.');
        }
        throw e;
      }

      event.locals.user = authUser as unknown as Record<string, any>;

      if (authUser.must_change_password && !isMustChangePasswordAllowedPath(pathname, method)) {
        if (pathname.startsWith('/api/')) {
          return jsonError(403);
        }
        return redirectTo('/settings/account');
      }

      if (!canAccessProtectedRoute(authUser, pathname, method)) {
        if (pathname.startsWith('/api/')) {
          return jsonError(403);
        }
        return redirectTo('/');
      }

      if (pathname.startsWith('/api/') && CSRF_MUTATION_METHODS.has(method)) {
        const origin = event.request.headers.get('Origin');
        const referer = event.request.headers.get('Referer');
        const eventOrigin = event.url.origin;
        const allowed = (origin && eventOrigin === getBaseFromUrl(origin))
          || (referer && eventOrigin === getBaseFromUrl(referer));
        if (!allowed) {
          return jsonError(403, 'csrf validation failed');
        }
      }
    }

    return resolve(event);
  }

  if (accessToken && !isAccessTokenValid(accessToken) && refreshToken && hasRefreshToken(refreshToken)) {
    const pair = await tryRefresh(event, event.fetch);
    if (pair) {
      event.locals.accessToken = pair.access_token;
      const refreshedUser = getUserFromToken(pair.access_token);
      event.locals.user = refreshedUser as unknown as Record<string, any> | undefined;

      if (refreshedUser) {
        try {
          const valid = await apiValidateAuthWithFetch(event.fetch, pair.access_token);
          if (!valid) {
            clearCookies(event);
            if (pathname.startsWith('/api/')) {
              return jsonError(401, 'unauthorized');
            }
            return redirectTo(`/login?from=${encodeURIComponent(pathname)}`);
          }
        } catch (e) {
          if (e instanceof AuthValidationUnavailableError) {
            return jsonError(503, 'Layanan validasi sesi sedang bermasalah. Silakan coba beberapa saat lagi.');
          }
          throw e;
        }

        if (!canAccessProtectedRoute(refreshedUser, pathname, method)) {
          if (pathname.startsWith('/api/')) {
            return jsonError(403);
          }
          return redirectTo('/');
        }
      }

      return resolve(event);
    }
  }

  clearCookies(event);

  if (pathname.startsWith('/api/')) {
    return jsonError(401, 'unauthorized');
  }

  return redirectTo(`/login?from=${encodeURIComponent(pathname)}`);
};

function getBaseFromUrl(urlStr: string): string {
  try {
    return new URL(urlStr).origin;
  } catch {
    return '';
  }
}
