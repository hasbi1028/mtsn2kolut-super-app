import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

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
		const body = await readRequestJson<Record<string, unknown>>(event.request);
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

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/students/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'students PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		await proxy(event).del(apiPath`/api/students/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'students DELETE');
	}
};

export const PATCH = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const status = body.status;
		if (!status) return json({ error: 'status wajib diisi' }, { status: 400 });
		const data = await proxy(event).patch(apiPath`/api/students/${id}/lifecycle`, { status });
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'students PATCH');
	}
};
