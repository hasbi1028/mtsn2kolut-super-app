import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const { id } = event.params;
		const data = await proxy(event).get(`/api/cbt/questions/${id}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/[id] GET');
	}
};

export const PUT = async (event: RequestEvent) => {
	try {
		const { id } = event.params;
		const body = await event.request.json();
		const data = await proxy(event).put(`/api/cbt/questions/${id}`, body);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/[id] PUT');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const { id } = event.params;
		await proxy(event).del(`/api/cbt/questions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions/[id] DELETE');
	}
};
