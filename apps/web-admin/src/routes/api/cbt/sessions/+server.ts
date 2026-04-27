import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, apiPost, apiDelete, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async () => {
	try {
		const data = await apiGet('/api/cbt/sessions');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions GET');
	}
};

export const POST: RequestHandler = async ({ request }) => {
	try {
		const body = await request.json() as Record<string, unknown>;
		const { package_id, class_id, title, scheduled_start, scheduled_end, status } = body;
		if (!package_id || !class_id || !title || !scheduled_start || !scheduled_end) {
			return json({ error: 'package_id, class_id, title, scheduled_start, scheduled_end wajib diisi' }, { status: 400 });
		}
		const data = await apiPost('/api/cbt/sessions', {
			package_id, class_id, title, scheduled_start, scheduled_end,
			status: status ?? 'draft',
		});
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions POST');
	}
};

export const DELETE: RequestHandler = async ({ url }) => {
	try {
		const id = url.searchParams.get('id');
		if (!id) return json({ error: 'id required' }, { status: 400 });
		await apiDelete(`/api/cbt/sessions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions DELETE');
	}
};
