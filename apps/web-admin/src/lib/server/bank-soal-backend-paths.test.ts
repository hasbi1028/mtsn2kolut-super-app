import { describe, expect, it } from 'vitest';

import { bankSoalApiPath, bankSoalBackendPath, bankSoalBackendPathWithQuery } from './bank-soal-backend-paths';

describe('bank-soal backend path helpers', () => {
	it('builds only canonical Bank Soal backend paths', () => {
		expect(bankSoalBackendPath('/questions')).toBe('/api/bank-soal/questions');
		expect(bankSoalBackendPath('/questions/123')).toBe('/api/bank-soal/questions/123');
		expect(bankSoalApiPath`/questions/${'abc 123'}`).toBe('/api/bank-soal/questions/abc%20123');
		expect(bankSoalBackendPath('/assets/asset-1/file')).toBe('/api/bank-soal/assets/asset-1/file');
		expect(bankSoalBackendPath('/soal-support/subjects')).toBe('/api/bank-soal/soal-support/subjects');
	});

	it('keeps assessment and unsafe paths out of the Bank Soal helper', () => {
		for (const path of ['/events', '/sessions', '/questions/../users', '/questions?limit=1', '/questions#frag', '/questions//asset']) {
			expect(() => bankSoalBackendPath(path), path).toThrow();
		}
	});

	it('adds query strings through the shared API query helper', () => {
		expect(bankSoalBackendPathWithQuery('/questions', new URLSearchParams({ limit: '10' }))).toBe(
			'/api/bank-soal/questions?limit=10'
		);
	});
});
