import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';
import { optionalStringField, requireStringField, validationErrorResponse } from '$lib/server/validation';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(apiPath`/api/parents/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'parents/[id] GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/parents/${id}`, {
			nama: requireStringField(body, 'nama'),
			phone: optionalStringField(body, 'phone'),
			address: optionalStringField(body, 'address'),
		});
		return json(data);
	} catch (e) {
		const validation = validationErrorResponse(e);
		if (validation) return validation;
		return handleRouteError(e, 'parents/[id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(apiPath`/api/parents/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'parents/[id] DELETE');
	}
};
