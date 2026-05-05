import { redirect, type RequestHandler } from '@sveltejs/kit';

const MODE_ROUTES = {
	catalog: '/cbt/bank-soal',
	composer: '/cbt/soal',
	import: '/cbt/soal/import',
	review: '/cbt/soal/review'
} as const;

type RetainedMode = keyof typeof MODE_ROUTES;

function retainedMode(value: string | null): RetainedMode | null {
	return value === 'catalog' || value === 'composer' || value === 'import' || value === 'review'
		? value
		: null;
}

export const GET: RequestHandler = ({ url }) => {
	const mode = retainedMode(url.searchParams.get('mode'));
	const target = new URL(MODE_ROUTES[mode ?? 'composer'], url);
	const questionId = url.searchParams.get('question_id');
	const eventId = url.searchParams.get('event_id');

	if (questionId) target.searchParams.set('question_id', questionId);
	if (eventId) target.searchParams.set('event_id', eventId);

	throw redirect(307, `${target.pathname}${target.search}`);
};
