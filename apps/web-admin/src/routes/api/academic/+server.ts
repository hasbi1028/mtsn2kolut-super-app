import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/academic');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const entity = event.url.searchParams.get('entity');
		if (!entity) return json({ error: 'entity required' }, { status: 400 });
		const body = await event.request.json() as Record<string, unknown>;
		const data = await proxy(event).post(`/api/academic/${entity}`, body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'academic POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const entity = event.url.searchParams.get('entity');
		const id = event.url.searchParams.get('id');
		if (!entity || !id) return json({ error: 'entity and id required' }, { status: 400 });
		await proxy(event).del(`/api/academic/${entity}/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'academic DELETE');
	}
};
