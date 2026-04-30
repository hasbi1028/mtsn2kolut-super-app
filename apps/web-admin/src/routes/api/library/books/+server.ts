import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const search = event.url.searchParams.get('search') ?? '';
		const kategori = event.url.searchParams.get('kategori') ?? '';
		const data = await proxy(event).get(`/api/library/books?search=${encodeURIComponent(search)}&kategori=${encodeURIComponent(kategori)}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'library/books GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const body = await event.request.json();
		const data = await proxy(event).post('/api/library/books', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'library/books POST');
	}
};
