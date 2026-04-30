import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

function canAccessJournal(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin') || roles.includes('guru');
}

export const POST = async (event: RequestEvent) => {
	if (!canAccessJournal(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = event.params.id;
		const body = await event.request.json() as Record<string, unknown>;
		const data = await proxy(event).post(`/api/journal/sessions/${id}/attendances`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'journal/sessions/[id]/attendances POST');
	}
};
