import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

function canAccessJournal(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin') || roles.includes('guru');
}

function isAdmin(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin');
}

export const GET = async (event: RequestEvent) => {
	if (!canAccessJournal(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = event.params.id;
		const data = await proxy(event).get(`/api/journal/sessions/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'journal/sessions/[id] GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	if (!canAccessJournal(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = event.params.id;
		const body = await event.request.json() as Record<string, unknown>;
		const data = await proxy(event).put(`/api/journal/sessions/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'journal/sessions/[id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	if (!isAdmin(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const id = event.params.id;
		await proxy(event).del(`/api/journal/sessions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'journal/sessions/[id] DELETE');
	}
};
