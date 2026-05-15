import { cbtApiPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy, readRequestJson } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	const id = event.params.id ?? '';
	try {
		const data = await proxy(event).get(cbtApiPath`/packages/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/packages/[id] GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	const id = event.params.id ?? '';
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(cbtApiPath`/packages/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/packages/[id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	const id = event.params.id ?? '';
	try {
		await proxy(event).del(cbtApiPath`/packages/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/packages/[id] DELETE');
	}
};
