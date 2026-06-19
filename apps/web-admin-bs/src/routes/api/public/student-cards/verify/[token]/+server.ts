import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const token = (event.params as Record<string, string | undefined>).token ?? '';
		const data = await proxy(event).get(`/api/public/student-cards/verify/${encodeURIComponent(token)}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'public student card verify');
	}
};
