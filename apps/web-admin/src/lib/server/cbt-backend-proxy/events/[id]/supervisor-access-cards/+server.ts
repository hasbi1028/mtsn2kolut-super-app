import { cbtBackendPathWithQuery } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readOptionalRequestJson, requiredRouteParam } from '$lib/server/api';

function cardPath(event: RequestEvent, suffix = '') {
	const id = requiredRouteParam(event.params.id, 'id');
	return cbtBackendPathWithQuery(`/events/${id}/supervisor-access-cards${suffix}`, event.url.searchParams);
}

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(cardPath(event));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/supervisor-access-cards GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readOptionalRequestJson<Record<string, unknown>>(event.request, {});
		const data = await proxy(event).post(cardPath(event, '/issue'), body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/events/[id]/supervisor-access-cards/issue POST');
	}
};
