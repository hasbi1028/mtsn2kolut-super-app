import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		const result = await proxy(event).get<any>(`/api/academic/rombels/${id}`);
		return new Response(JSON.stringify(result), {
			status: 200,
			headers: { 'content-type': 'application/json' },
		});
	} catch (e) {
		return handleRouteError(e, 'rombel GET');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.params.id;
		await proxy(event).del(`/api/academic/rombels/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'rombels DELETE');
	}
};
