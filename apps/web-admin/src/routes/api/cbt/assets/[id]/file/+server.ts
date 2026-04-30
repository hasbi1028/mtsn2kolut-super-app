import { env } from '$env/dynamic/private';
import type { RequestEvent } from '@sveltejs/kit';

const BASE = (env.API_BASE_URL ?? 'http://localhost:8080').replace(/\/$/, '');

// No auth required — asset file endpoint is public on the backend.
// UUID provides sufficient access control for exam media assets.
export const GET = async (event: RequestEvent) => {
	const upstream = await fetch(`${BASE}/api/cbt/assets/${event.params.id}/file`);
	if (!upstream.ok) {
		return new Response(null, { status: upstream.status });
	}

	const contentType = upstream.headers.get('content-type') ?? 'application/octet-stream';
	const contentLength = upstream.headers.get('content-length');
	const responseHeaders: Record<string, string> = {
		'content-type': contentType,
		'cache-control': 'public, max-age=3600, immutable',
	};
	if (contentLength) responseHeaders['content-length'] = contentLength;

	return new Response(upstream.body, { status: 200, headers: responseHeaders });
};
