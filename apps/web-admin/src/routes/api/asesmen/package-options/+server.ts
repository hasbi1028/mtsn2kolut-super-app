import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const subjectId = event.url.searchParams.get('subject_id') ?? '';
		const suffix = subjectId ? `?subject_id=${encodeURIComponent(subjectId)}` : '';
		const data = await proxy(event).get(`${apiPath`/api/asesmen/package-options`}${suffix}`);
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'asesmen/package-options GET');
	}
};
