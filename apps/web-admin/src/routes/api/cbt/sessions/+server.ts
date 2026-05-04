import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

function queryPath(path: string, params: URLSearchParams) {
	const query = params.toString();
	return query ? `${path}?${query}` : path;
}

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(queryPath('/api/cbt/sessions', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
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
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		await proxy(event).del(apiPath`/api/cbt/sessions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions DELETE');
	}
};
