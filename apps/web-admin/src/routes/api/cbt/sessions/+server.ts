import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/cbt/sessions');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json() as Record<string, unknown>;
		const { package_id, title, scheduled_start, scheduled_end, status, scope_type, class_id } = body;
		if (!package_id || !title || !scheduled_start || !scheduled_end) {
			return json({ error: 'package_id, title, scheduled_start, scheduled_end wajib diisi' }, { status: 400 });
		}
		if ((scope_type ?? 'class') === 'class' && !class_id) {
			return json({ error: 'class_id wajib diisi untuk scope_type=class' }, { status: 400 });
		}
		const data = await proxy(event).post('/api/cbt/sessions', { ...body, status: status ?? 'draft' });
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.url.searchParams.get('id');
		if (!id) return json({ error: 'id required' }, { status: 400 });
		await proxy(event).del(`/api/cbt/sessions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions DELETE');
	}
};
