import { json } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const INTERNAL_KEY = env.INTERNAL_API_KEY ?? '';

export class ApiError extends Error {
	constructor(public status: number, message: string) { super(message); }
}

function headers(): Record<string, string> {
	const h: Record<string, string> = { 'Content-Type': 'application/json' };
	h['X-Internal-Key'] = INTERNAL_KEY;
	return h;
}

async function unwrap<T>(res: Response): Promise<T> {
	if (res.status === 204) return null as T;
	const json = await res.json() as { data?: T; error?: string };
	if (!res.ok) throw new ApiError(res.status, json.error ?? `HTTP ${res.status}`);
	return json.data as T;
}

export async function apiGet<T>(path: string): Promise<T> {
	return unwrap<T>(await fetch(`${BASE}${path}`, { headers: headers() }));
}

export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		method: 'POST', headers: headers(),
		body: body !== undefined ? JSON.stringify(body) : undefined,
	});
	return unwrap<T>(res);
}

export async function apiPut<T>(path: string, body?: unknown): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		method: 'PUT', headers: headers(),
		body: body !== undefined ? JSON.stringify(body) : undefined,
	});
	return unwrap<T>(res);
}

export async function apiDelete<T>(path: string, body?: unknown): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		method: 'DELETE', headers: headers(),
		body: body !== undefined ? JSON.stringify(body) : undefined,
	});
	return unwrap<T>(res);
}

export type TokenPair = {
	access_token: string;
	refresh_token: string;
};

// Public login endpoint — returns access + refresh JWT tokens
export async function apiLogin(username: string, password: string): Promise<TokenPair> {
	const res = await fetch(`${BASE}/api/auth/login`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ username, password }),
	});
	const json = await res.json() as { data?: TokenPair; error?: string };
	if (!res.ok) throw new ApiError(res.status, json.error ?? 'unauthorized');
	return json.data!;
}

export async function apiRefresh(refresh_token: string): Promise<TokenPair> {
	const res = await fetch(`${BASE}/api/auth/refresh`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
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
