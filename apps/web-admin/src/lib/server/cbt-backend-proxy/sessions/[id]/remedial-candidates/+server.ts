import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const query = event.url.searchParams.toString();
		const base = cbtApiPath`/sessions/${id}/remedial-candidates`;
		const data = await proxy(event).get(query ? `${base}?${query}` : base);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/sessions remedial-candidates GET');
	}
};
