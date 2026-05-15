import { cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

function queryPath(path: string, params: URLSearchParams) {
	const query = params.toString();
	return query ? `${path}?${query}` : path;
}

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(queryPath(cbtBackendPath('/packages/readiness'), event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/packages/readiness GET');
	}
};
