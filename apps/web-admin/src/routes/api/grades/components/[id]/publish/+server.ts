import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

function canManageGrades(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin') || roles.includes('guru');
}

export const PATCH = async (event: RequestEvent) => {
	if (!canManageGrades(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = event.params.id;
		const body = await event.request.json() as Record<string, unknown>;
		const data = await proxy(event).patch(`/api/grades/components/${id}/publish`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'grades/components/[id]/publish PATCH');
	}
};
