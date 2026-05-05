import { describe, expect, it } from 'vitest';

import { GET } from './+server';

function expectRedirect(url: string, location: string) {
	try {
		GET({ url: new URL(url) } as never);
	} catch (error) {
		expect(error).toMatchObject({ status: 307, location });
		return;
	}
	throw new Error('expected redirect');
}

describe('/cbt/questions retired route', () => {
	it('redirects legacy question-bank visits to the canonical composer route', () => {
		expectRedirect('http://localhost/cbt/questions', '/cbt/soal');
	});

	it('preserves safe legacy question id and retained mode query parameters', () => {
		expectRedirect(
			'http://localhost/cbt/questions?question_id=abc-123&mode=review&unsafe=ignored',
			'/cbt/soal/review?question_id=abc-123'
		);
	});

	it('maps retained import mode to the import page route', () => {
		expectRedirect(
			'http://localhost/cbt/questions?mode=import&event_id=evt-1',
			'/cbt/soal/import?event_id=evt-1'
		);
	});

	it('drops retired experiment modes during redirect', () => {
		expectRedirect(
			'http://localhost/cbt/questions?question_id=abc-123&mode=studio',
			'/cbt/soal?question_id=abc-123'
		);
	});
});
