import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

type ApiEnvelope<T> = {
	data?: T;
	error?: string;
};

async function readProxyResponse<T>(response: Response, fallback: string): Promise<{ data?: T; error?: string }> {
	const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | null;
	if (!response.ok) {
		return { error: payload?.error ?? fallback };
	}
	if (!payload || payload.data === undefined) {
		return { error: fallback };
	}
	return { data: payload.data };
}

export const GET: RequestHandler = async (event) => {
	try {
		const params = new URLSearchParams();
		for (const key of ['search', 'category_id', 'status', 'classification_code']) {
			const value = event.url.searchParams.get(key) ?? '';
			if (value) params.set(key, value);
		}
		const suffix = params.toString() ? `?${params.toString()}` : '';
		const data = await proxy(event).get(`/api/tu/archives/documents${suffix}`);
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
		const result = await readProxyResponse<unknown>(response, 'Gagal mengunggah arsip.');
		if (result.error) return json({ error: result.error }, { status: response.status || 502 });
		return json(result.data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'tu/archives/documents POST');
	}
};
