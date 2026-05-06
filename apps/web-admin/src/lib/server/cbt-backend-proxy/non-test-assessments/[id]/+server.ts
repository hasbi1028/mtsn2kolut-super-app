import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).get(cbtApiPath`/non-test-assessments/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/non-test-assessments/[id] GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(cbtApiPath`/non-test-assessments/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/non-test-assessments/[id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		await proxy(event).del(cbtApiPath`/non-test-assessments/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/non-test-assessments/[id] DELETE');
	}
};
