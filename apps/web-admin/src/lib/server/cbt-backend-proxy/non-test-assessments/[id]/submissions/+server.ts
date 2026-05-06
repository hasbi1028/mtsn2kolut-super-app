import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(cbtApiPath`/non-test-assessments/${id}/submissions`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/non-test-assessments/[id]/submissions GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(cbtApiPath`/non-test-assessments/${id}/submissions`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/non-test-assessments/[id]/submissions POST');
	}
};
