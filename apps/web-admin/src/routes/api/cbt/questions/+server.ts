import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { apiGet, apiPost, apiDelete, handleRouteError } from '$lib/server/api';

export const GET: RequestHandler = async () => {
	try {
		const data = await apiGet('/api/cbt/questions');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions GET');
	}
};

export const POST: RequestHandler = async ({ request }) => {
	try {
		const body = await request.json() as Record<string, unknown>;
		const { subject_id, code, question_text, option_a, option_b, option_c, option_d, option_e, answer_key, explanation, difficulty, status } = body;
		if (!subject_id || !question_text || !option_a || !option_b || !option_c || !option_d || !answer_key) {
			return json({ error: 'subject_id, question_text, option A-D, dan answer_key wajib diisi' }, { status: 400 });
		}
		const data = await apiPost('/api/cbt/questions', {
			subject_id, code: code ?? '', question_text,
			option_a, option_b, option_c, option_d, option_e: option_e ?? '',
			answer_key, explanation: explanation ?? '',
			difficulty: difficulty ?? 'medium', status: status ?? 'draft',
		});
		return json(data, { status: 201 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions POST');
	}
};

export const DELETE: RequestHandler = async ({ url }) => {
	try {
		const id = url.searchParams.get('id');
		if (!id) return json({ error: 'id required' }, { status: 400 });
		await apiDelete(`/api/cbt/questions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions DELETE');
	}
};
