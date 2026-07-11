import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const search = event.url.searchParams.get('search') ?? '';
		const status = event.url.searchParams.get('status') ?? '';
		const templateCode = event.url.searchParams.get('template_code') ?? '';
		const params = new URLSearchParams();
		if (search) params.set('search', search);
		if (status) params.set('status', status);
		if (templateCode) params.set('template_code', templateCode);
		const data = await proxy(event).get(apiPathWithQuery('/api/tu/surat-keterangan', params));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat-keterangan GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/tu/surat-keterangan', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'tu/surat-keterangan POST');
	}
};
