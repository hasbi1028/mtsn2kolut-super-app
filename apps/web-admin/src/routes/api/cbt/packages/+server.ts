import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/cbt/packages');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/packages GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json() as Record<string, unknown>;
		const { subject_id, title, description, duration_minutes, randomize_questions, is_active, question_ids } = body;
		if (!subject_id || !title || !duration_minutes) {
			return json({ error: 'subject_id, title, duration_minutes wajib diisi' }, { status: 400 });
		}
		const data = await proxy(event).post('/api/cbt/packages', {
			subject_id, title, description: description ?? '',
			duration_minutes: Number(duration_minutes),
			randomize_questions: randomize_questions ?? false,
			is_active: is_active ?? true,
			question_ids: question_ids ?? [],
		});
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/packages POST');
	}
};

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.url.searchParams.get('id');
		if (!id) return json({ error: 'id required' }, { status: 400 });
		await proxy(event).del(`/api/cbt/packages/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/packages DELETE');
	}
};
