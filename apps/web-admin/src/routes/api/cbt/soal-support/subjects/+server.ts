import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

type AcademicPayload = {
	subjects?: unknown[];
};

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get<AcademicPayload>('/api/academic');
		return json({ subjects: Array.isArray(data.subjects) ? data.subjects : [] });
	} catch (e) {
		return handleRouteError(e, 'cbt/soal-support/subjects GET');
	}
};
