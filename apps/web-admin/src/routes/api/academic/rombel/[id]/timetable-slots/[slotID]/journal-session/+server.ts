import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

type OpenJournalSessionResponse = {
	created?: boolean;
};

function journalAccessError(event: RequestEvent): Response | null {
	const user = event.locals.user;
	if (!user) return json({ error: 'unauthorized' }, { status: 401 });
	const roles = user.roles ?? (user.role ? [user.role] : []);
	if (roles.includes('admin') || roles.includes('guru')) return null;
	return json({ error: 'forbidden' }, { status: 403 });
}

export const POST = async (event: RequestEvent) => {
	const accessError = journalAccessError(event);
	if (accessError) return accessError;

	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const slotID = requiredRouteParam(event.params.slotID, 'slotID');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post<OpenJournalSessionResponse>(
			apiPath`/api/academic/rombel/${id}/timetable-slots/${slotID}/journal-session`,
			body
		);
		return json(data, { status: data.created ? 201 : 200 });
	} catch (e) {
		return handleRouteError(e, 'academic rombel timetable journal-session POST');
	}
};
