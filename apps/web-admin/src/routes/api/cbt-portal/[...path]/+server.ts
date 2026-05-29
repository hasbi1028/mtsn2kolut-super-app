import { env } from '$env/dynamic/private';
import type { RequestEvent } from '@sveltejs/kit';
import { readLimitedRequestText, RequestPayloadError } from '$lib/server/api';

const API_BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const BODY_LIMIT_BYTES = 96 * 1024;

function upstreamPath(event: RequestEvent): string {
	const path = event.params.path?.trim() ?? '';
	const suffix = path ? `/${path.split('/').map(encodeURIComponent).join('/')}` : '';
	return `/api/cbt-portal${suffix}${event.url.search}`;
}

function upstreamHeaders(event: RequestEvent): Headers {
	const headers = new Headers();
	for (const key of ['accept', 'content-type', 'user-agent', 'authorization']) {
		const value = event.request.headers.get(key);
		if (value) headers.set(key, value);
	}
	try {
		const clientIP = event.getClientAddress().trim();
		if (clientIP) {
			headers.set('x-forwarded-for', clientIP);
			headers.set('x-real-ip', clientIP);
		}
	} catch {
		// ignore adapters that do not expose client address
	}
	return headers;
}

async function requestBody(request: Request): Promise<BodyInit | undefined> {
	if (request.method === 'GET' || request.method === 'HEAD') return undefined;
	try {
		return await readLimitedRequestText(request, BODY_LIMIT_BYTES);
	} catch (error) {
		if (error instanceof RequestPayloadError) {
			throw new Response(JSON.stringify({ error: 'Payload ujian terlalu besar' }), {
				status: 413,
				headers: { 'Content-Type': 'application/json' }
			});
		}
		throw error;
	}
}


type JSONValue = null | boolean | number | string | JSONValue[] | { [key: string]: JSONValue };

const SENSITIVE_PORTAL_RESPONSE_KEYS = new Set([
	'answer_key',
	'answerkey',
	'correct_answer',
	'correctanswer',
	'correct_answers',
	'correctanswers',
	'kunci_jawaban',
	'kuncijawaban',
	'pin',
	'raw_pin',
	'raw_token'
]);

function redactPortalResponse(value: JSONValue): JSONValue {
	if (Array.isArray(value)) return value.map(redactPortalResponse);
	if (!value || typeof value !== 'object') return value;
	const redacted: { [key: string]: JSONValue } = {};
	for (const [key, child] of Object.entries(value)) {
		if (SENSITIVE_PORTAL_RESPONSE_KEYS.has(key.replace(/[-_\s]/g, '').toLowerCase())) continue;
		redacted[key] = redactPortalResponse(child);
	}
	return redacted;
}

async function responseBody(res: Response): Promise<BodyInit | null> {
	const contentType = res.headers.get('content-type') ?? '';
	if (!contentType.toLowerCase().includes('application/json')) return res.body;
	const text = await res.text();
	if (!text) return text;
	try {
		return JSON.stringify(redactPortalResponse(JSON.parse(text) as JSONValue));
	} catch {
		return text;
	}
}

async function forwardCbtPortal(event: RequestEvent): Promise<Response> {
	try {
		const res = await event.fetch(`${API_BASE}${upstreamPath(event)}`, {
			method: event.request.method,
			headers: upstreamHeaders(event),
			body: await requestBody(event.request),
			redirect: 'manual'
		});
		const headers = new Headers();
		const contentType = res.headers.get('content-type');
		if (contentType) headers.set('content-type', contentType);
		headers.set('cache-control', 'no-store');
		return new Response(await responseBody(res), { status: res.status, headers });
	} catch (error) {
		if (error instanceof Response) return error;
		return new Response(JSON.stringify({ error: 'Layanan portal ujian sedang bermasalah' }), {
			status: 502,
			headers: { 'Content-Type': 'application/json' }
		});
	}
}

export const GET = forwardCbtPortal;
export const POST = forwardCbtPortal;
export const OPTIONS = () => new Response(null, { status: 204 });
