import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, apiPathWithQuery, handleRouteError, proxy, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get(apiPathWithQuery('/api/bank-soal/reviewer-scopes', event.url.searchParams));
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'bank-soal/reviewer-scopes GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).post('/api/bank-soal/reviewer-scopes', body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'bank-soal/reviewer-scopes POST');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const data = await proxy(event).put('/api/bank-soal/reviewer-scopes', body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'bank-soal/reviewer-scopes PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = requiredRouteParam(event.url.searchParams.get('id') ?? undefined, 'id');
		await proxy(event).del(apiPath`/api/bank-soal/reviewer-scopes/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'bank-soal/reviewer-scopes DELETE');
	}
};
