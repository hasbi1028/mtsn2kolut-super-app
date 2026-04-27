const BASE = (process.env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const INTERNAL_KEY = process.env.INTERNAL_API_KEY ?? '';
const WORKER_KEY   = process.env.WORKER_API_KEY   ?? '';

export class ApiError extends Error {
	constructor(public status: number, message: string) { super(message); }
}

function headers(isWorker = false): Record<string, string> {
	const h: Record<string, string> = { 'Content-Type': 'application/json' };
	if (isWorker) h['X-Worker-Key'] = WORKER_KEY;
	else          h['X-Internal-Key'] = INTERNAL_KEY;
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

// Worker proxy — returns raw Response for caller to handle 204 / format adapt
export async function workerFetch(method: 'GET' | 'POST', path: string, workerIdHeader?: string, body?: unknown): Promise<Response> {
	const h = headers(true);
	if (workerIdHeader) h['x-worker-id'] = workerIdHeader;
	return fetch(`${BASE}${path}`, {
		method,
		headers: h,
		body: body !== undefined ? JSON.stringify(body) : undefined,
	});
}

// Public login endpoint — returns JWT token
export async function apiLogin(username: string, password: string): Promise<string> {
	const res = await fetch(`${BASE}/api/auth/login`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ username, password }),
	});
	const json = await res.json() as { data?: { token: string }; error?: string };
	if (!res.ok) throw new ApiError(res.status, json.error ?? 'unauthorized');
	return json.data!.token;
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
