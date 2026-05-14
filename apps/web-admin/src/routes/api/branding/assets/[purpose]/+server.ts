import type { RequestHandler } from './$types';
import { apiPath, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const purpose = requiredRouteParam(event.params.purpose, 'purpose');
		const form = await event.request.formData();
		const response = await proxy(event).fetch(apiPath`/api/branding/assets/${purpose}`, {
			method: 'POST',
			body: form
		});
		return Response.json(await response.json(), { status: response.status });
	} catch (e) {
		return handleRouteError(e, 'POST /api/branding/assets/[purpose]');
	}
};

export const DELETE: RequestHandler = async (event) => {
	try {
		const purpose = requiredRouteParam(event.params.purpose, 'purpose');
		return Response.json(await proxy(event).del(apiPath`/api/branding/assets/${purpose}`));
	} catch (e) {
		return handleRouteError(e, 'DELETE /api/branding/assets/[purpose]');
	}
};
