import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const DELETE = async (event: RequestEvent) => {
	try {
		await proxy(event).del(`/api/cbt/sessions/${event.params.id}/rooms/${event.params.rid}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions/[id]/rooms/[rid] DELETE');
	}
};
