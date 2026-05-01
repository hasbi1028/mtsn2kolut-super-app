import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post('/api/document-cycles/obligations/generate-year', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'document-cycles/obligations/generate-year POST');
	}
};
