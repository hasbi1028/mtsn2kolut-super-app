import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

function canManageGrades(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin') || roles.includes('guru');
}

export const POST = async (event: RequestEvent) => {
	if (!canManageGrades(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(apiPath`/api/grades/assignments/${id}/finalize`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'grades/assignments/[id]/finalize POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	if (!canManageGrades(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(apiPath`/api/grades/assignments/${id}/finalize`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'grades/assignments/[id]/finalize DELETE');
	}
};
