import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPathWithQuery, proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const search = event.url.searchParams.get('search') ?? '';
		const status = event.url.searchParams.get('status') ?? '';
		const params = new URLSearchParams();
		if (search) params.set('search', search);
		if (status) params.set('status', status);
		const data = await proxy(event).get(apiPathWithQuery('/api/tu/surat-keterangan/students', params));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/surat-keterangan/students GET');
	}
};
