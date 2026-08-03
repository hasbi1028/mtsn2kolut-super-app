import type { RequestEvent } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

/**
 * Passthrough worker API (tanpa sesi admin).
 *
 * Worker DomCloud memakai BACKEND_URL https://mtsn2kolut.sch.id (domain
 * publik → BFF ini). Request worker (/api/pusaka/worker/claim|heartbeat|
 * config|complete|fail|...) harus diteruskan ke core-api dengan
 * Authorization header ASLI (Bearer WORKER_API_KEY) — bukan sesi browser.
 *
 * Keamanan: route ini murni proxy; auth tetap divalidasi core-api
 * (WORKER_API_KEY). Tanpa key → 401 dari core-api. Route khusus
 * (status, restart) tetap menang atas catch-all ini.
 */

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

async function forward(event: RequestEvent): Promise<Response> {
	const path = (event.params as Record<string, string | undefined>).path ?? '';
	if (path.includes('..') || path.includes('//')) {
		return new Response(JSON.stringify({ error: 'invalid path' }), { status: 400 });
	}

	const url = `${BASE}/api/pusaka/worker/${path}${event.url.search}`;
	const headers = new Headers();

	// Forward Authorization asli (Bearer worker key) — jangan pakai sesi admin.
	const authorization = event.request.headers.get('authorization');
	if (authorization) headers.set('authorization', authorization);
	const contentType = event.request.headers.get('content-type');
	if (contentType) headers.set('content-type', contentType);
	const accept = event.request.headers.get('accept');
	if (accept) headers.set('accept', accept);

	const isBodyless = ['GET', 'HEAD', 'OPTIONS'].includes(event.request.method.toUpperCase());
	const upstream = await event.fetch(url, {
		method: event.request.method,
		headers,
		body: isBodyless ? undefined : await event.request.arrayBuffer(),
	});

	// Forward respons mentah (status + body) tanpa unwrap — worker butuh bentuk asli.
	return new Response(upstream.body, {
		status: upstream.status,
		headers: {
			'content-type': upstream.headers.get('content-type') ?? 'application/json',
		},
	});
}

export const GET = forward;
export const POST = forward;
export const PUT = forward;
export const PATCH = forward;
export const DELETE = forward;
export const OPTIONS = forward;
