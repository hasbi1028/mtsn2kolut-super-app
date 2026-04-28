import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const DELETE = async (event: RequestEvent) => {
	try {
		await proxy(event).del(`/api/users/${event.params.id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'users DELETE');
	}
};
