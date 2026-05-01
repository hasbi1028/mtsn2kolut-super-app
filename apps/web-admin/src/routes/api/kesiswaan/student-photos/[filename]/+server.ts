import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const response = await proxy(event).fetch(`/api/kesiswaan/student-photos/${event.params.filename}`);
		if (!response.ok) return new Response(null, { status: response.status });
		return new Response(response.body, {
			headers: {
				'content-type': response.headers.get('content-type') ?? 'application/octet-stream',
				'cache-control': 'private, max-age=3600',
			},
		});
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/student-photos/[filename] GET');
	}
};
