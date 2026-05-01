import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const code = event.url.searchParams.get('classification_code') ?? '';
		const tanggal = event.url.searchParams.get('tanggal') ?? '';
		const data = await proxy(event).get(
			`/api/tu/surat/outgoing/preview-number?classification_code=${encodeURIComponent(code)}&tanggal=${encodeURIComponent(tanggal)}`
		);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat/outgoing/preview-number GET');
	}
};
