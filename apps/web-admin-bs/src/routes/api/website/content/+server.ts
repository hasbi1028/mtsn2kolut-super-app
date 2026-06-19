import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPathWithQuery, handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const qs = new URLSearchParams();
		for (const key of ['kind', 'status', 'search']) {
			const value = event.url.searchParams.get(key);
			if (value) qs.set(key, value);
		}
		const items = await proxy(event).get(apiPathWithQuery('/api/website/content', qs));
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'website/content GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const item = await proxy(event).post('/api/website/content', body);
		return json(item, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'website/content POST');
	}
};
