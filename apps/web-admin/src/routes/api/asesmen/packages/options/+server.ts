import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET: RequestHandler = async (event) => {
	try {
		const subjectID = event.url.searchParams.get('subject_id') ?? '';
		const suffix = subjectID ? `?subject_id=${encodeURIComponent(subjectID)}` : '';
		const data = await proxy(event).get(`/api/asesmen/packages/options${suffix}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/packages/options GET');
	}
};
