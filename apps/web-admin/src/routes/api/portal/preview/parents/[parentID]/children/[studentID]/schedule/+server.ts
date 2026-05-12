import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const parentID = event.params.parentID ?? '';
		const studentID = event.params.studentID ?? '';
		const data = await proxy(event).get(apiPath`/api/portal/preview/parents/${parentID}/children/${studentID}/schedule`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'portal/preview/parents/[parentID]/children/[studentID]/schedule GET');
	}
};
