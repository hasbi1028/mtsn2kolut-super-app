import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async ({ params }) => {
	try {
		const data = await apiGet(`/api/cbt/sessions/${params.id}/results`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/sessions results GET');
	}
};
