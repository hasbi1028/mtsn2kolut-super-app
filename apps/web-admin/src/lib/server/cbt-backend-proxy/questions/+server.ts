import { cbtApiPath, cbtBackendPath } from '$lib/server/cbt-backend-paths';
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import {
	apiPath,
	apiPathWithQuery,
	handleRouteError,
	proxy,
	readRequestJson,
	requiredRouteParam
} from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery(cbtBackendPath('/questions'), event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const { subject_id } = body;
		if (!subject_id) {
			return json({ error: 'subject_id wajib diisi' }, { status: 400 });
		}
		const data = await proxy(event).post(cbtBackendPath('/questions'), body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		await proxy(event).del(cbtApiPath`/questions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions DELETE');
	}
};
