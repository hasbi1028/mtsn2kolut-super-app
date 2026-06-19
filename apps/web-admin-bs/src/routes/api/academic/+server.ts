import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import {
	apiPath,
	handleRouteError,
	proxy,
	readRequestJson,
	requiredRouteParam
} from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/academic');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const entity = requiredRouteParam(event.url.searchParams.get('entity') ?? undefined, 'entity');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post(apiPath`/api/academic/${entity}`, body);
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'academic POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const entity = requiredRouteParam(event.url.searchParams.get('entity') ?? undefined, 'entity');
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		await proxy(event).del(apiPath`/api/academic/${entity}/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'academic DELETE');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const entity = requiredRouteParam(event.url.searchParams.get('entity') ?? undefined, 'entity');
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put(apiPath`/api/academic/${entity}/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'academic PUT');
	}
};
