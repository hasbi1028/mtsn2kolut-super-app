import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';
import { optionalStringField, requireStringField, validationErrorResponse } from '$lib/server/validation';

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
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const nama = requireStringField(body, 'nama');
		const phone = optionalStringField(body, 'phone');
		const address = optionalStringField(body, 'address');
		const data = await proxy(event).post('/api/parents', {
			nama,
			phone,
			address,
		});
		return json(data, { status: 201 });
	} catch (e) {
		const validation = validationErrorResponse(e);
		if (validation) return validation;
		return handleRouteError(e, 'parents POST');
	}
};
