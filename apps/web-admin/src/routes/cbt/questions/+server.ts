import { redirect, type RequestHandler } from '@sveltejs/kit';

const RETAINED_MODES = new Set(['catalog', 'composer', 'review', 'import']);

export const GET: RequestHandler = ({ url }) => {
	const target = new URL('/cbt/soal', url);
	const questionId = url.searchParams.get('question_id');
	const mode = url.searchParams.get('mode');

	if (questionId) target.searchParams.set('question_id', questionId);
	if (mode && RETAINED_MODES.has(mode)) target.searchParams.set('mode', mode);

	throw redirect(307, `${target.pathname}${target.search}`);
};
