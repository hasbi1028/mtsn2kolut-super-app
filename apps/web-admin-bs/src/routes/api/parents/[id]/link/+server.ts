import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';
import { requireUuidField, validationErrorResponse } from '$lib/server/validation';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(apiPath`/api/parents/${id}/link`, {
			student_id: requireUuidField(body, 'student_id'),
		});
		return json(data);
	} catch (e) {
		const validation = validationErrorResponse(e);
		if (validation) return validation;
		return handleRouteError(e, 'parents/[id]/link POST');
	}
};
