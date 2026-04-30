import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

function canAccessJournal(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin') || roles.includes('guru');
}

export const GET = async (event: RequestEvent) => {
	if (!canAccessJournal(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const params = event.url.searchParams.toString();
		const path = params ? `/api/journal?${params}` : '/api/journal';
		const data = await proxy(event).get(path);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'journal GET');
	}
};
