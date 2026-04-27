import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, apiPost, apiDelete, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async () => {
	try {
		const data = await apiGet('/api/academic');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic GET');
	}
};

export const POST: RequestHandler = async ({ request, url }) => {
	try {
		const entity = url.searchParams.get('entity');
		if (!entity) return json({ error: 'entity required' }, { status: 400 });
		const body = await request.json() as Record<string, unknown>;
		const data = await apiPost(`/api/academic/${entity}`, body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'academic POST');
	}
};

export const DELETE: RequestHandler = async ({ url }) => {
	try {
		const entity = url.searchParams.get('entity');
		const id = url.searchParams.get('id');
		if (!entity || !id) return json({ error: 'entity and id required' }, { status: 400 });
		await apiDelete(`/api/academic/${entity}/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'academic DELETE');
	}
};
