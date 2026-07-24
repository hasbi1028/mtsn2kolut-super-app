import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		await proxy(event).del(`/api/academic/curriculum/allocations/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'allocation DELETE');
	}
};
