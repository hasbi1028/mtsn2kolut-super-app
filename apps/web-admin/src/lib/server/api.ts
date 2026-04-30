import { json } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import type { RequestEvent } from '@sveltejs/kit';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const INTERNAL_KEY = env.INTERNAL_API_KEY ?? '';

export class ApiError extends Error {
	constructor(public status: number, message: string) { super(message); }
}

export function requireAuthorizationHeader(accessToken?: string): Record<string, string> {
	if (!accessToken) {
		throw new ApiError(401, 'unauthorized');
	}
	return {
		Authorization: `Bearer ${accessToken}`,
	};
}

export function requireAuthHeaders(accessToken?: string): Record<string, string> {
	return {
		'Content-Type': 'application/json',
		...requireAuthorizationHeader(accessToken),
	};
}

export function publicHeaders(): Record<string, string> {
	return { 'Content-Type': 'application/json' };
}

export function internalHeaders(): Record<string, string> {
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
	if (INTERNAL_KEY) {
		headers['X-Internal-Key'] = INTERNAL_KEY;
	}
	return headers;
}

function headers(bearerToken?: string): Record<string, string> {
	return bearerToken ? requireAuthHeaders(bearerToken) : publicHeaders();
}

async function unwrap<T>(res: Response): Promise<T> {
	if (res.status === 204) return null as T;
	const json = await res.json() as { data?: T; error?: string };
	if (!res.ok) throw new ApiError(res.status, json.error ?? `HTTP ${res.status}`);
	return json.data as T;
}

export async function apiGet<T>(path: string, accessToken?: string): Promise<T> {
	return unwrap<T>(await fetch(`${BASE}${path}`, { headers: headers(accessToken) }));
}

export async function apiPost<T>(path: string, body?: unknown, accessToken?: string): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		method: 'POST', headers: headers(accessToken),
		body: body !== undefined ? JSON.stringify(body) : undefined,
	});
	return unwrap<T>(res);
}

export async function apiPut<T>(path: string, body?: unknown, accessToken?: string): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		method: 'PUT', headers: headers(accessToken),
		body: body !== undefined ? JSON.stringify(body) : undefined,
	});
	return unwrap<T>(res);
}

export async function apiPatch<T>(path: string, body?: unknown, accessToken?: string): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		method: 'PATCH', headers: headers(accessToken),
		body: body !== undefined ? JSON.stringify(body) : undefined,
	});
	return unwrap<T>(res);
}

export async function apiDelete<T>(path: string, body?: unknown, accessToken?: string): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		method: 'DELETE', headers: headers(accessToken),
		body: body !== undefined ? JSON.stringify(body) : undefined,
	});
	return unwrap<T>(res);
}

export function proxy(event: RequestEvent) {
	const accessToken = event.locals.accessToken ?? event.cookies.get('access_token');
	if (!accessToken) {
		throw new ApiError(401, 'unauthorized');
	}
	return {
		get: <T>(path: string) => apiGet<T>(path, accessToken),
		post: <T>(path: string, body?: unknown) => apiPost<T>(path, body, accessToken),
		put: <T>(path: string, body?: unknown) => apiPut<T>(path, body, accessToken),
		patch: <T>(path: string, body?: unknown) => apiPatch<T>(path, body, accessToken),
		del: <T>(path: string, body?: unknown) => apiDelete<T>(path, body, accessToken),
		fetch: (path: string, init?: RequestInit) => {
			const headers = requireAuthorizationHeader(accessToken);
			return fetch(`${BASE}${path}`, { ...init, headers: { ...headers, ...(init?.headers as Record<string, string>) } });
		},
	};
}

export type TokenPair = {
	access_token: string;
	refresh_token: string;
};

type ClientMeta = {
	userAgent?: string;
	ipAddress?: string;
};

function clientMetaHeaders(meta?: ClientMeta): Record<string, string> {
	const headers: Record<string, string> = {};
	if (meta?.userAgent) headers['X-Client-User-Agent'] = meta.userAgent;
	if (meta?.ipAddress) headers['X-Client-IP'] = meta.ipAddress;
	return headers;
}

// Public login endpoint — returns access + refresh JWT tokens
export async function apiLogin(username: string, password: string, meta?: ClientMeta): Promise<TokenPair> {
	const res = await fetch(`${BASE}/api/auth/login`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json', ...clientMetaHeaders(meta) },
		body: JSON.stringify({ username, password }),
	});
	const json = await res.json() as { data?: TokenPair; error?: string };
	if (!res.ok) throw new ApiError(res.status, json.error ?? 'unauthorized');
	return json.data!;
}

export async function apiRefresh(refresh_token: string, meta?: ClientMeta): Promise<TokenPair> {
	const res = await fetch(`${BASE}/api/auth/refresh`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json', ...clientMetaHeaders(meta) },
		body: JSON.stringify({ refresh_token }),
	});
	const json = await res.json() as { data?: TokenPair; error?: string };
	if (!res.ok) throw new ApiError(res.status, json.error ?? 'unauthorized');
	return json.data!;
}

// Error handler untuk semua route proxy — kembalikan JSON error yang jelas
export function handleRouteError(e: unknown, route = ''): Response {
	if (e instanceof ApiError) {
		const prefix = route ? `[${route}]` : '[api-proxy]';
		console.error(`${prefix} upstream ${e.status}:`, e.message);
		// Jangan expose detail 5xx dari Go sebagai 500 SvelteKit — pakai 503
		const status = e.status >= 500 ? 503 : e.status;
		return json({ error: e.message }, { status });
	}
	const msg = (e as Error)?.message ?? String(e);
	const route_label = route ? `[${route}]` : '[api-proxy]';
	console.error(`${route_label} network/unhandled error:`, msg);
	return json({ error: 'Backend tidak dapat dihubungi — pastikan Go API berjalan' }, { status: 503 });
}

// Format a Timestamptz ISO string to WITA display string
export function toWITA(iso: string | null | undefined): string {
	if (!iso) return '';
	const d = new Date(iso);
	return new Intl.DateTimeFormat('sv-SE', {
		timeZone: 'Asia/Makassar', year: 'numeric', month: '2-digit',
		day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit',
	}).format(d).replace('T', ' ') + ' WITA';
}
