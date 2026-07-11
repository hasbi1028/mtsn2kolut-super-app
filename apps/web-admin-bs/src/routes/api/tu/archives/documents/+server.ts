import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, jsonProxyResponse, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const params = new URLSearchParams();
		for (const key of ['search', 'category_id', 'status', 'classification_code']) {
			const value = event.url.searchParams.get(key) ?? '';
			if (value) params.set(key, value);
		}
		const data = await proxy(event).get(apiPathWithQuery('/api/tu/archives/documents', params));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'tu/archives/documents GET');
	}
};

export const POST: RequestHandler = async (event) => {
	try {
		const form = await event.request.formData();
		const response = await proxy(event).fetch('/api/tu/archives/documents', {
			method: 'POST',
			body: form,
		});
		return await jsonProxyResponse<{ data?: unknown }, unknown>(response, {
			fallbackMessage: 'Gagal mengunggah arsip.',
			status: 201,
			map: (result) => result.data ?? result
		});
	} catch (e) {
		return handleRouteError(e, 'tu/archives/documents POST');
	}
};
