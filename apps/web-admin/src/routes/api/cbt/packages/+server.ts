import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, apiPost, apiDelete, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async () => {
	try {
		const data = await apiGet('/api/cbt/packages');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/packages GET');
	}
};

export const POST: RequestHandler = async ({ request }) => {
	try {
		const body = await request.json() as Record<string, unknown>;
		const { subject_id, title, description, duration_minutes, randomize_questions, is_active, question_ids } = body;
		if (!subject_id || !title || !duration_minutes) {
			return json({ error: 'subject_id, title, duration_minutes wajib diisi' }, { status: 400 });
		}
		const data = await apiPost('/api/cbt/packages', {
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

export const DELETE: RequestHandler = async ({ url }) => {
	try {
		const id = url.searchParams.get('id');
		if (!id) return json({ error: 'id required' }, { status: 400 });
		await apiDelete(`/api/cbt/packages/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/packages DELETE');
	}
};
