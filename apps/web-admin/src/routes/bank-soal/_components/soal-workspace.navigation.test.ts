import { describe, expect, it } from 'vitest';

import {
	bankSoalListHref,
	bankSoalQuestionDetailHref,
	composerRouteQuestionId,
	detailRouteQuestionId,
} from './soal-workspace.navigation';

describe('Bank Soal composer route navigation helpers', () => {
	it('keeps detail route path id from being overwritten by query question_id', () => {
		const params = new URLSearchParams({
			question_id: 'query-id',
			event_id: 'event-1',
		});

		expect(detailRouteQuestionId('path-id')).toBe('path-id');
		expect(composerRouteQuestionId('path-id', params)).toBe('path-id');
		expect(composerRouteQuestionId('', params)).toBe('query-id');
	});

	it('preserves only safe list context when building detail and back hrefs', () => {
		const params = new URLSearchParams({
			event_id: 'event 1',
			subject_id: 'subject/1',
			workflow_status: 'review',
			question_id: 'query-id',
			redirect: 'https://example.test',
		});

		expect(bankSoalListHref(params)).toBe('/bank-soal?event_id=event+1&subject_id=subject%2F1&workflow_status=review');
		expect(bankSoalQuestionDetailHref('path/id', params)).toBe('/bank-soal/soal/path%2Fid?event_id=event+1&subject_id=subject%2F1&workflow_status=review');
	});
});
