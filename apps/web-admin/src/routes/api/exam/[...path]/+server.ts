import { env } from '$env/dynamic/private';
import type { RequestEvent } from '@sveltejs/kit';

const API_BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const EXAM_BODY_LIMIT_BYTES = 96 * 1024;

function examPath(event: RequestEvent): string {
	const path = event.params.path?.trim() ?? '';
	const suffix = path ? `/${path.split('/').map(encodeURIComponent).join('/')}` : '';
	return `/api/exam${suffix}${event.url.search}`;
}

function upstreamHeaders(request: Request): Headers {
	const headers = new Headers();
	for (const key of ['accept', 'content-type', 'user-agent', 'x-exam-token', 'x-device-fingerprint', 'x-forwarded-for', 'x-real-ip']) {
		const value = request.headers.get(key);
		if (value) headers.set(key, value);
	}
	return headers;
}

async function requestBody(request: Request): Promise<BodyInit | undefined> {
	if (request.method === 'GET' || request.method === 'HEAD') return undefined;
	const text = await request.text();
	if (text.length > EXAM_BODY_LIMIT_BYTES) {
		throw new Response(JSON.stringify({ error: 'Payload ujian terlalu besar' }), {
			status: 413,
			headers: { 'Content-Type': 'application/json' }
		});
	}
	return text;
}

async function forwardExam(event: RequestEvent): Promise<Response> {
	try {
		const res = await event.fetch(`${API_BASE}${examPath(event)}`, {
			method: event.request.method,
			headers: upstreamHeaders(event.request),
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
		return new Response(JSON.stringify({ error: 'Layanan ujian sedang bermasalah' }), {
			status: 502,
			headers: { 'Content-Type': 'application/json' }
		});
	}
}

export const GET = forwardExam;
export const POST = forwardExam;
export const OPTIONS = () => new Response(null, { status: 204 });
