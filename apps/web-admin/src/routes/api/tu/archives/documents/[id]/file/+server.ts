import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const response = await proxy(event).fetch(`/api/tu/archives/documents/${event.params.id}/file`);
		if (!response.ok) {
			const payload = (await response.json().catch(() => null)) as { error?: string } | null;
			return json({ error: payload?.error ?? 'Gagal mengambil file arsip.' }, { status: response.status });
		}
		const headers = new Headers();
		for (const key of ['content-type', 'content-length', 'content-disposition', 'cache-control']) {
			const value = response.headers.get(key);
			if (value) headers.set(key, value);
		}
		return new Response(response.body, { status: response.status, headers });
	} catch (e) {
		return handleRouteError(e, 'tu/archives/documents/[id]/file GET');
	}
};
