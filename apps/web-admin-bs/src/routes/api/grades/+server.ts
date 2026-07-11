import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, proxy, handleRouteError } from '$lib/server/api';

function canManageGrades(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin') || roles.includes('guru');
}

export const GET = async (event: RequestEvent) => {
	if (!canManageGrades(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/grades', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'grades GET');
	}
};
