import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/parents');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'parents GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json() as Record<string, unknown>;
		const { nama, phone, address } = body;
		if (!nama) {
			return json({ error: 'nama wajib diisi' }, { status: 400 });
		}
		const data = await proxy(event).post('/api/parents', {
			nama,
			phone: phone ?? '',
			address: address ?? '',
		});
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'parents POST');
	}
};
