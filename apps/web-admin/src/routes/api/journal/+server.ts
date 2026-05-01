import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, proxy, handleRouteError } from '$lib/server/api';

function canAccessJournal(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin') || roles.includes('guru');
}

export const GET = async (event: RequestEvent) => {
	if (!canAccessJournal(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/journal', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'journal GET');
	}
};
