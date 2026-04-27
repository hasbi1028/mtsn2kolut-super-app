import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, apiPost, apiDelete, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async () => {
	try {
		const data = await apiGet('/api/students');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'students GET');
	}
};

export const POST: RequestHandler = async ({ request }) => {
	try {
		const body = await request.json() as Record<string, unknown>;
		const { nis, nisn, nama, gender, parent_name, parent_phone, class_id, is_active } = body;
		if (!nis || !nama || !gender) {
			return json({ error: 'nis, nama, gender wajib diisi' }, { status: 400 });
		}
		const data = await apiPost('/api/students', {
			nis, nisn: nisn ?? '', nama, gender, parent_name: parent_name ?? '',
			parent_phone: parent_phone ?? '', class_id: class_id ?? '', is_active: is_active ?? true,
		});
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'students POST');
	}
};

export const DELETE: RequestHandler = async ({ url }) => {
	try {
		const id = url.searchParams.get('id');
		if (!id) return json({ error: 'id required' }, { status: 400 });
		await apiDelete(`/api/students/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'students DELETE');
	}
};
