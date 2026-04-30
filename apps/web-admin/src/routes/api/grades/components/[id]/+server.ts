import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

function canManageGrades(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin') || roles.includes('guru');
}

export const DELETE = async (event: RequestEvent) => {
	if (!canManageGrades(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = event.params.id;
		await proxy(event).del(`/api/grades/components/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'grades/components/[id] DELETE');
	}
};

export const PUT = async (event: RequestEvent) => {
	if (!canManageGrades(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = event.params.id;
		const body = await event.request.json() as Record<string, unknown>;
		const data = await proxy(event).put(`/api/grades/components/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'grades/components/[id] PUT');
	}
};
