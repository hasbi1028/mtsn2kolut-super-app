import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { handleRouteError, proxy } from '$lib/server/api';

export interface Rombel {
	id: string;
	code: string;
	name: string;
	level: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
	academic_year_id: string;
	academic_year_name: string;
}

export const GET = async (event: RequestEvent) => {
	try {
		const items = await proxy(event).get<Rombel[]>('/api/academic/rombels');
		return json({ items });
	} catch (e) {
		return handleRouteError(e, 'rombels GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json();
		const result = await proxy(event).post<Rombel>('/api/academic/rombels', body);
		return json(result, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'rombels POST');
	}
};
