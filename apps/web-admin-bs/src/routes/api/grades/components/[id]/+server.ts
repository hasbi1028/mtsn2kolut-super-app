import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

function canManageGrades(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin') || roles.includes('guru');
}

export const DELETE = async (event: RequestEvent) => {
	if (!canManageGrades(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(apiPath`/api/grades/components/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'grades/components/[id] DELETE');
	}
};

export const PUT = async (event: RequestEvent) => {
	if (!canManageGrades(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/grades/components/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'grades/components/[id] PUT');
	}
};
