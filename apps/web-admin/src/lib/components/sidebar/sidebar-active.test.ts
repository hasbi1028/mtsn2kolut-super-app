import { describe, expect, it } from 'vitest';
import { findActiveSidebarHref, matchesSidebarPath } from './sidebar-active';

const items = [
	{ href: '/' },
	{ href: '/bank-soal' },
	{ href: '/bank-soal/analisis-butir' }
];

describe('sidebar active route matching', () => {
	it('uses segment-aware matching', () => {
		expect(matchesSidebarPath('/bank-soal/daftar', '/bank-soal')).toBe(true);
		expect(matchesSidebarPath('/bank-soal-arsip', '/bank-soal')).toBe(false);
	});

	it('selects only the most specific active href to avoid double highlights', () => {
		expect(findActiveSidebarHref('/bank-soal', items)).toBe('/bank-soal');
		expect(findActiveSidebarHref('/bank-soal/daftar', items)).toBe('/bank-soal');
		expect(findActiveSidebarHref('/bank-soal/daftar/abc', items)).toBe('/bank-soal');
		expect(findActiveSidebarHref('/bank-soal/analisis-butir', items)).toBe('/bank-soal/analisis-butir');
	});

});
