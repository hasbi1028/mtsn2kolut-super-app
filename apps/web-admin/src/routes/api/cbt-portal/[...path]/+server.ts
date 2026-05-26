import { env } from '$env/dynamic/private';
import type { RequestEvent } from '@sveltejs/kit';

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
	const text = await request.text();
	if (new TextEncoder().encode(text).byteLength > BODY_LIMIT_BYTES) {
		throw new Response(JSON.stringify({ error: 'Payload ujian terlalu besar' }), {
			status: 413,
			headers: { 'Content-Type': 'application/json' }
		});
	}
	return text;
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
		return new Response(res.body, { status: res.status, headers });
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
