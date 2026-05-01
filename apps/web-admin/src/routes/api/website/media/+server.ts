import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, jsonProxyResponse, proxy } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const formData = await event.request.formData();
		const res = await proxy(event).fetch('/api/website/media', {
			method: 'POST',
			body: formData,
		});

		return await jsonProxyResponse<Record<string, unknown>>(res, {
			fallbackMessage: 'Gagal upload gambar.',
			status: 201
		});
	} catch (e) {
		return handleRouteError(e, 'website/media POST');
	}
};
