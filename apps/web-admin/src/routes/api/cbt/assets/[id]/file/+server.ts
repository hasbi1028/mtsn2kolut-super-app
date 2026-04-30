import { handleRouteError, proxy } from '$lib/server/api';
import type { RequestEvent } from '@sveltejs/kit';

export const GET = async (event: RequestEvent) => {
	try {
		const upstream = await proxy(event).fetch(`/api/cbt/assets/${event.params.id}/file${event.url.search}`);
		if (!upstream.ok) {
			return new Response(null, { status: upstream.status });
		}

		const contentType = upstream.headers.get('content-type') ?? 'application/octet-stream';
		const contentLength = upstream.headers.get('content-length');
		const contentDisposition = upstream.headers.get('content-disposition');
		const cacheControl = upstream.headers.get('cache-control') ?? 'private, max-age=300';
		const responseHeaders: Record<string, string> = {
			'content-type': contentType,
			'cache-control': cacheControl,
		};
		if (contentLength) responseHeaders['content-length'] = contentLength;
		if (contentDisposition) responseHeaders['content-disposition'] = contentDisposition;

		return new Response(upstream.body, { status: 200, headers: responseHeaders });
	} catch (e) {
		return handleRouteError(e, 'cbt/assets/[id]/file GET');
	}
};
