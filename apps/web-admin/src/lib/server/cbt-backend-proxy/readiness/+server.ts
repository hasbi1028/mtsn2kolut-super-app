import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const query = event.url.searchParams.toString();
		const path = query ? `${cbtApiPath`/readiness`}?${query}` : cbtApiPath`/readiness`;
		const data = await proxy(event).get(path);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/readiness GET');
	}
};
