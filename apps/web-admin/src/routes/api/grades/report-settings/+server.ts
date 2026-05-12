import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

function canManageGrades(event: RequestEvent) {
	const roles = event.locals.user?.roles ?? (event.locals.user?.role ? [event.locals.user.role] : []);
	return roles.includes('admin') || roles.includes('guru');
}

export const GET = async (event: RequestEvent) => {
	if (!canManageGrades(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const data = await proxy(event).get('/api/grades/report-settings');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'grades report-settings GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	if (!canManageGrades(event)) return json({ error: 'forbidden' }, { status: 403 });
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put('/api/grades/report-settings', body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'grades report-settings PUT');
	}
};
