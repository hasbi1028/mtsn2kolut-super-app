import { describe, expect, it } from 'vitest';
import { safeSameOriginRedirectPath } from './redirects';

describe('safeSameOriginRedirectPath', () => {
	it('accepts same-origin relative paths beginning with slash', () => {
		expect(safeSameOriginRedirectPath('/dashboard?tab=1')).toBe('/dashboard?tab=1');
		expect(safeSameOriginRedirectPath('%2Fparents%3Fq%3Dortu')).toBe('/parents?q=ortu');
	});

	it('rejects open redirect and scheme-like values', () => {
		for (const value of ['https://evil.test', 'http://evil.test', '//evil.test/path', 'javascript:alert(1)', 'dashboard', '/\\evil.test']) {
			expect(safeSameOriginRedirectPath(value)).toBe('/');
		}
	});

	it('falls back for login loops and control characters', () => {
		for (const value of ['/login', '/login?from=%2Fdashboard', '/login#form', '/dashboard\nSet-Cookie:bad=1']) {
			expect(safeSameOriginRedirectPath(value)).toBe('/');
		}
	});
});
