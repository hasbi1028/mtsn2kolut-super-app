import { json } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import type { RequestEvent } from '@sveltejs/kit';
import type { AuthUser } from '$lib/server/auth';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const GENERIC_UPSTREAM_ERROR = 'Layanan backend sedang bermasalah. Silakan coba beberapa saat lagi.';
const INVALID_UPSTREAM_JSON = 'Respons backend kosong atau bukan JSON';
export const AUTH_VALIDATION_PROBE_PATH = '/api/auth/sessions';
type Fetcher = typeof fetch;

export class ApiError extends Error {
	constructor(
		public status: number,
		message: string,
		public upstreamMessage = message
	) {
		super(message);
	}
}

export class RequestPayloadError extends Error {
	constructor(message = 'Payload JSON tidak valid') {
		super(message);
	}
}

export class AuthValidationUnavailableError extends Error {
	constructor(message = 'Layanan validasi sesi belum tersedia') {
		super(message);
	}
}

type ApiEnvelope<T> = {
	data?: T;
	error?: string;
	message?: string;
};

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null;
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

export async function readRequestJson<T>(request: Request): Promise<T> {
	try {
		return await request.json() as T;
	} catch (e) {
		if (isMalformedJsonError(e)) throw new RequestPayloadError();
		throw e;
	}
}

export async function readOptionalRequestJson<T>(request: Request, fallback: T): Promise<T> {
	const text = await request.text();
	if (!text.trim()) return fallback;
	try {
		return JSON.parse(text) as T;
	} catch {
		throw new RequestPayloadError();
	}
}

function extractMessage(payload: unknown): string | undefined {
	if (!payload || typeof payload !== 'object') return undefined;
	const envelope = payload as ApiEnvelope<unknown>;
	return envelope.error ?? envelope.message;
}

function headersToRecord(headers: HeadersInit | undefined): Record<string, string> {
	const record: Record<string, string> = {};
	new Headers(headers).forEach((value, key) => {
		record[key] = value;
	});
	return record;
}

function withoutProxyControlledHeaders(headers: Record<string, string>): Record<string, string> {
	return Object.fromEntries(
		Object.entries(headers).filter(([key]) => {
			const normalized = key.toLowerCase();
			return normalized !== 'x-internal-key' && normalized !== 'authorization';
		})
	);
}

function isMalformedJsonError(e: unknown): boolean {
	return e instanceof SyntaxError && /json|unexpected end|unexpected token|not valid json/i.test(e.message);
}

async function parseEnvelope<T>(res: Response): Promise<ApiEnvelope<T> | undefined> {
	const text = await res.text();
	if (!text) return undefined;
	try {
		return JSON.parse(text) as ApiEnvelope<T>;
	} catch {
		return undefined;
	}
}

async function unwrap<T>(res: Response): Promise<T> {
	if (res.status === 204) return null as T;
	const payload = await parseEnvelope<T>(res);
	if (!res.ok) {
		const upstreamMessage = extractMessage(payload);
		const safeMessage = res.status >= 500 ? GENERIC_UPSTREAM_ERROR : upstreamMessage ?? `HTTP ${res.status}`;
		throw new ApiError(res.status, safeMessage, upstreamMessage ?? `HTTP ${res.status}`);
	}
	if (!isRecord(payload)) {
		throw new ApiError(502, GENERIC_UPSTREAM_ERROR, INVALID_UPSTREAM_JSON);
	}
	const upstreamError = extractMessage(payload);
	if (upstreamError) {
		throw new ApiError(502, GENERIC_UPSTREAM_ERROR, upstreamError);
	}
	if (!('data' in payload)) {
		throw new ApiError(502, GENERIC_UPSTREAM_ERROR, INVALID_UPSTREAM_JSON);
	}
	return (payload as ApiEnvelope<T>).data as T;
}

export async function readProxyJson<T>(res: Response, fallbackMessage?: string): Promise<T> {
	if (res.status === 204) return null as T;
	const payload = await parseEnvelope<unknown>(res) as T | undefined;
	if (!res.ok) {
		const upstreamMessage = extractMessage(payload) ?? fallbackMessage ?? `HTTP ${res.status}`;
		const safeMessage = res.status >= 500 ? GENERIC_UPSTREAM_ERROR : upstreamMessage;
		throw new ApiError(res.status, safeMessage, upstreamMessage);
	}
	if (payload === undefined || payload === null) {
		throw new ApiError(502, GENERIC_UPSTREAM_ERROR, fallbackMessage ?? INVALID_UPSTREAM_JSON);
	}
	return payload as T;
}

type JsonProxyOptions<T, R> = {
	fallbackMessage?: string;
	status?: number;
	map?: (payload: T) => R;
};

export async function jsonProxyResponse<T, R = T>(
	res: Response,
	options: JsonProxyOptions<T, R> = {}
): Promise<Response> {
	const payload = await readProxyJson<T>(res, options.fallbackMessage);
	const status = options.status ?? res.status;
	if (status === 204) {
		return new Response(null, { status: 204 });
	}
	const body = options.map ? options.map(payload) : payload;
	return json(body, { status });
}

type StreamProxyOptions = {
	fallbackMessage?: string;
	defaultContentType?: string;
	defaultCacheControl?: string;
	headers?: readonly string[];
};

const DEFAULT_STREAM_HEADERS = ['content-type', 'content-length', 'content-disposition', 'cache-control', 'x-content-type-options'] as const;

export async function streamProxyResponse(res: Response, options: StreamProxyOptions = {}): Promise<Response> {
	if (!res.ok) {
		await readProxyJson(res, options.fallbackMessage);
		throw new ApiError(res.status, options.fallbackMessage ?? `HTTP ${res.status}`);
	}

	const headers = new Headers();
	for (const key of options.headers ?? DEFAULT_STREAM_HEADERS) {
		const value = res.headers.get(key);
		if (value) headers.set(key, value);
	}
	if (!headers.has('content-type')) {
		headers.set('content-type', options.defaultContentType ?? 'application/octet-stream');
	}
	if (options.defaultCacheControl && !headers.has('cache-control')) {
		headers.set('cache-control', options.defaultCacheControl);
	}
	if (!headers.has('x-content-type-options')) {
		headers.set('x-content-type-options', 'nosniff');
	}
	return new Response(res.body, { status: res.status, headers });
}

async function apiRequest<T>(
	fetcher: Fetcher,
	path: string,
	init: RequestInit,
	headers: Record<string, string>
): Promise<T> {
	const res = await fetcher(`${BASE}${path}`, {
		...init,
		headers: {
			...headersToRecord(init.headers),
			...headers,
		},
	});
	return unwrap<T>(res);
}

function jsonBody(body: unknown): BodyInit | undefined {
	return body !== undefined ? JSON.stringify(body) : undefined;
}

export function apiPath(strings: TemplateStringsArray, ...values: Array<string | number | boolean>): string {
	let path = strings[0] ?? '';
	for (let i = 0; i < values.length; i += 1) {
		path += encodeURIComponent(String(values[i]));
		path += strings[i + 1] ?? '';
	}
	return path;
}

export function apiPathWithQuery(path: string, params: URLSearchParams | string): string {
	const query = typeof params === 'string' ? params.replace(/^\?/, '') : params.toString();
	return query ? `${path}?${query}` : path;
}

export function requiredRouteParam(value: string | undefined, name: string): string {
	if (!value) {
		throw new ApiError(400, `${name} tidak valid`);
	}
	return value;
}

export async function apiPublicGetWithFetch<T>(fetcher: Fetcher, path: string): Promise<T> {
	return apiRequest<T>(fetcher, path, {}, publicHeaders());
}

export async function apiPublicPostWithFetch<T>(fetcher: Fetcher, path: string, body?: unknown): Promise<T> {
	return apiRequest<T>(fetcher, path, {
		method: 'POST',
		body: jsonBody(body),
	}, publicHeaders());
}

export function createApiClient(event: RequestEvent) {
	const currentAccessToken = () => event.locals.accessToken ?? event.cookies.get('access_token');
	requireAuthorizationHeader(currentAccessToken());

	return {
		get: <T>(path: string) => apiRequest<T>(event.fetch, path, {}, requireAuthHeaders(currentAccessToken())),
		post: <T>(path: string, body?: unknown) => apiRequest<T>(event.fetch, path, {
			method: 'POST',
			body: jsonBody(body),
		}, requireAuthHeaders(currentAccessToken())),
		put: <T>(path: string, body?: unknown) => apiRequest<T>(event.fetch, path, {
			method: 'PUT',
			body: jsonBody(body),
		}, requireAuthHeaders(currentAccessToken())),
		patch: <T>(path: string, body?: unknown) => apiRequest<T>(event.fetch, path, {
			method: 'PATCH',
			body: jsonBody(body),
		}, requireAuthHeaders(currentAccessToken())),
		del: <T>(path: string, body?: unknown) => apiRequest<T>(event.fetch, path, {
			method: 'DELETE',
			body: jsonBody(body),
		}, requireAuthHeaders(currentAccessToken())),
		fetch: (path: string, init?: RequestInit) => {
			const headers = requireAuthorizationHeader(currentAccessToken());
			const initHeaders = withoutProxyControlledHeaders(headersToRecord(init?.headers));
			return event.fetch(`${BASE}${path}`, { ...init, headers: { ...initHeaders, ...headers } });
		},
	};
}

export function proxy(event: RequestEvent) {
	return createApiClient(event);
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

export async function apiLoginWithFetch(
	fetcher: Fetcher,
	username: string,
	password: string,
	meta?: ClientMeta
): Promise<TokenPair> {
	return apiRequest<TokenPair>(fetcher, '/api/auth/login', {
		method: 'POST',
		body: JSON.stringify({ username, password }),
	}, { 'Content-Type': 'application/json', ...clientMetaHeaders(meta) });
}

export async function apiRefreshWithFetch(
	fetcher: Fetcher,
	refresh_token: string,
	meta?: ClientMeta
): Promise<TokenPair> {
	return apiRequest<TokenPair>(fetcher, '/api/auth/refresh', {
		method: 'POST',
		body: JSON.stringify({ refresh_token }),
	}, { 'Content-Type': 'application/json', ...clientMetaHeaders(meta) });
}

export async function apiValidateAuthWithFetch(fetcher: Fetcher, accessToken: string): Promise<boolean> {
	let res: Response;
	try {
		// Use a protected auth-domain read as the session-validity probe. The response body is
		// intentionally ignored; JWT/session liveness is enforced by the Go middleware.
		res = await fetcher(`${BASE}${AUTH_VALIDATION_PROBE_PATH}`, {
			headers: { Authorization: `Bearer ${accessToken}` },
		});
	} catch (e) {
		throw new AuthValidationUnavailableError((e as Error)?.message);
	}
	if (res.status === 401 || res.status === 403) return false;
	if (!res.ok) throw new AuthValidationUnavailableError(`HTTP ${res.status}`);
	return res.ok;
}

export async function getVerifiedUserFromAccessToken(
	fetcher: Fetcher,
	accessToken: string,
	decodeUser: (token: string) => AuthUser | null
): Promise<AuthUser | undefined> {
	if (!await apiValidateAuthWithFetch(fetcher, accessToken)) return undefined;
	return decodeUser(accessToken) ?? undefined;
}

// Error handler untuk semua route proxy — kembalikan JSON error yang jelas
export function handleRouteError(e: unknown, route = ''): Response {
	const route_label = route ? `[${route}]` : '[api-proxy]';
	if (e instanceof RequestPayloadError) {
		console.warn(`${route_label} malformed request JSON:`, e.message);
		return json({ error: e.message }, { status: 400 });
	}
	if (e instanceof ApiError) {
		const prefix = route ? `[${route}]` : '[api-proxy]';
		console.error(`${prefix} upstream ${e.status}:`, e.upstreamMessage);
		// Jangan expose detail 5xx dari Go sebagai 500 SvelteKit — pakai 503
		const status = e.status >= 500 ? 503 : e.status;
		const message = e.status >= 500 ? GENERIC_UPSTREAM_ERROR : e.message;
		return json({ error: message }, { status });
	}
	const msg = (e as Error)?.message ?? String(e);
	if (isMalformedJsonError(e)) {
		console.warn(`${route_label} malformed request JSON:`, msg);
		return json({ error: 'Payload JSON tidak valid' }, { status: 400 });
	}
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
