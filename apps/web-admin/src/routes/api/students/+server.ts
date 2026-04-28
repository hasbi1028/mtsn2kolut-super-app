import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/students');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'students GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json() as Record<string, unknown>;
		const { nis, nisn, nama, gender, parent_name, parent_phone, class_id, is_active } = body;
		if (!nis || !nama || !gender) {
			return json({ error: 'nis, nama, gender wajib diisi' }, { status: 400 });
		}
		const data = await proxy(event).post('/api/students', {
			nis, nisn: nisn ?? '', nama, gender, parent_name: parent_name ?? '',
			parent_phone: parent_phone ?? '', class_id: class_id ?? '', is_active: is_active ?? true,
		});
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'students POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.url.searchParams.get('id');
		if (!id) return json({ error: 'id required' }, { status: 400 });
		await proxy(event).del(`/api/students/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'students DELETE');
	}
};
