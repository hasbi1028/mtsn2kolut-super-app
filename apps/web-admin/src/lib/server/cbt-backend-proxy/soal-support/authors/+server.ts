import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

type AuthorsPayload = {
	items?: Array<{ username: string; display_name: string }>;
};

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get<AuthorsPayload>('/api/bank-soal/questions/authors');
		return json({ items: Array.isArray(data.items) ? data.items : [] });
	} catch (e) {
		return handleRouteError(e, 'cbt/soal-support/authors GET');
	}
};
