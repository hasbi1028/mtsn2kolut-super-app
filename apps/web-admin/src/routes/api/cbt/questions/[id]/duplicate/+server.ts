import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const { id } = event.params;
		const data = await proxy(event).post(`/api/cbt/questions/${id}/duplicate`);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/[id]/duplicate POST');
	}
};
