import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, requiredRouteParam } from '$lib/server/api';

export const POST = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const res = await proxy(event).post(apiPath`/api/users/${id}/force-password-change`, {});
		return json(res ?? { ok: true });
	} catch (e) {
		return handleRouteError(e, 'users force-password-change POST');
	}
};
