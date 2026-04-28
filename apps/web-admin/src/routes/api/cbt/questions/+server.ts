import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { proxy, handleRouteError } from '$lib/server/api';

export const GET = async (event: RequestEvent) => {
	try {
		const data = await proxy(event).get('/api/cbt/questions');
		return json(data);
	} catch (e) {
		return handleRouteError(e, 'cbt/questions GET');
	}
};

export const POST = async (event: RequestEvent) => {
	try {
		const body = await event.request.json() as Record<string, unknown>;
		const { subject_id, code, question_text, option_a, option_b, option_c, option_d, option_e, answer_key, explanation, difficulty, status } = body;
		if (!subject_id || !question_text || !option_a || !option_b || !option_c || !option_d || !answer_key) {
			return json({ error: 'subject_id, question_text, option A-D, dan answer_key wajib diisi' }, { status: 400 });
		}
		const data = await proxy(event).post('/api/cbt/questions', {
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

export const DELETE = async (event: RequestEvent) => {
	try {
		const id = event.url.searchParams.get('id');
		if (!id) return json({ error: 'id required' }, { status: 400 });
		await proxy(event).del(`/api/cbt/questions/${id}`);
		return new Response(null, { status: 204 });
	} catch (e) {
		return handleRouteError(e, 'cbt/questions DELETE');
	}
};
