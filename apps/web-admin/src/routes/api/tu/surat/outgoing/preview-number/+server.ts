import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPathWithQuery, proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/tu/surat/outgoing/preview-number', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat/outgoing/preview-number GET');
	}
};
