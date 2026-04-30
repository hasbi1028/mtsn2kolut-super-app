import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const qs = event.url.searchParams.toString();
		const data = await proxy(event).get(`/api/cbt/questions${qs ? `?${qs}` : ''}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json() as Record<string, unknown>;
		const { subject_id } = body;
		if (!subject_id) {
			return json({ error: 'subject_id wajib diisi' }, { status: 400 });
		}
		const data = await proxy(event).post('/api/cbt/questions', body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.url.searchParams.get('id');
		if (!id) return json({ error: 'id required' }, { status: 400 });
		await proxy(event).del(`/api/cbt/questions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions DELETE');
	}
};
