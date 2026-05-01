import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, jsonProxyResponse, requiredRouteParam } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const form = await event.request.formData();
		const response = await proxy(event).fetch(apiPath`/api/kesiswaan/students/${id}/photo`, {
			method: 'POST',
			body: form,
		});
		return await jsonProxyResponse<{ data?: unknown }, unknown>(response, {
			fallbackMessage: 'Gagal mengunggah foto siswa.',
			map: (payload) => payload.data ?? payload
		});
	} catch (e) {
		return handleRouteError(e, 'kesiswaan/students/[id]/photo POST');
	}
};
