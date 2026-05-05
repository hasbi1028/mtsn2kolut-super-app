import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

const MODE_ROUTES = {
	catalog: '/cbt/bank-soal',
	composer: '/cbt/soal',
	import: '/cbt/soal/import',
	review: '/cbt/soal/review'
} as const;

type LegacyMode = keyof typeof MODE_ROUTES;

function legacyMode(value: string | null): LegacyMode | null {
	return value === 'catalog' || value === 'composer' || value === 'import' || value === 'review'
		? value
		: null;
}

function legacyModeLocation(url: URL, mode: LegacyMode): string {
	const target = new URL(MODE_ROUTES[mode], url);
	const eventId = url.searchParams.get('event_id');
	const questionId = url.searchParams.get('question_id');

	if (eventId) target.searchParams.set('event_id', eventId);
	if (questionId && (mode === 'composer' || mode === 'review')) {
		target.searchParams.set('question_id', questionId);
	}

	return `${target.pathname}${target.search}`;
}

export const load: PageLoad = ({ url }) => {
	const mode = legacyMode(url.searchParams.get('mode'));
	if (!mode) return {};

	throw redirect(307, legacyModeLocation(url, mode));
};
