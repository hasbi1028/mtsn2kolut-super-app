import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const items = await proxy(event).get<any[]>('/api/academic/rombels/unassigned-students');
		return new Response(JSON.stringify({ items }), {
			status: 200,
			headers: { 'content-type': 'application/json' },
		});
	} catch (e) {
		return handleRouteError(e, 'unassigned students GET');
	}
};
