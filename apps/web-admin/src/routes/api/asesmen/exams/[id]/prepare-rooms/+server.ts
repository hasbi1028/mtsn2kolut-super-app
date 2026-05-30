import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy, requiredRouteParam } from '$lib/server/api';

export const POST: RequestHandler = async (event) => {
	try {
		const id = requiredRouteParam(event.params.id, 'id');
		const data = await proxy(event).post(apiPath`/api/asesmen/exams/${id}/prepare-rooms`, {});
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/exams/[id]/prepare-rooms POST');
	}
};
