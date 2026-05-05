import { describe, expect, it } from 'vitest';

import { load } from './+page';

function expectRedirect(url: string, location: string) {
	try {
		load({ url: new URL(url) } as never);
	} catch (error) {
		expect(error).toMatchObject({ status: 307, location });
		return;
	}
	throw new Error('expected redirect');
}

describe('/cbt/soal legacy mode redirects', () => {
	it('normalizes composer mode to the route page without mode query', () => {
		expectRedirect(
			'http://localhost/cbt/soal?mode=composer&question_id=abc-123',
			'/cbt/soal?question_id=abc-123'
		);
	});

	it('normalizes import mode to the import route page', () => {
		expectRedirect(
			'http://localhost/cbt/soal?mode=import&event_id=evt-1',
			'/cbt/soal/import?event_id=evt-1'
		);
	});

	it('normalizes catalog and review modes to their page routes', () => {
		expectRedirect('http://localhost/cbt/soal?mode=catalog', '/cbt/bank-soal');
		expectRedirect(
			'http://localhost/cbt/soal?mode=review&question_id=abc-123',
			'/cbt/soal/review?question_id=abc-123'
		);
	});

	it('allows the canonical composer route without query mode', () => {
		expect(load({ url: new URL('http://localhost/cbt/soal') } as never)).toEqual({});
	});
});
