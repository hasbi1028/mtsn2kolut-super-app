import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const classId = event.url.searchParams.get('class_id') || '';
		const items = await proxy(event).get<any[]>(`/api/academic/curriculum/assignments?class_id=${classId}`);
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'assignments GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const result = await proxy(event).post<any>('/api/academic/curriculum/assignments', body);
		return json(result, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'assignments POST');
	}
};
