import { describe, expect, it } from 'vitest';
import { findActiveSidebarHref, matchesSidebarPath } from './sidebar-active';

const items = [
	{ href: '/' },
	{ href: '/bank-soal' },
	{ href: '/bank-soal/daftar' },
	{ href: '/bank-soal/analisis-butir' },
	{ href: '/asesmen' },
	{ href: '/asesmen/kegiatan' },
	{ href: '/asesmen/kegiatan/new' }
];

describe('sidebar active route matching', () => {
	it('uses segment-aware matching', () => {
		expect(matchesSidebarPath('/bank-soal/daftar', '/bank-soal')).toBe(true);
		expect(matchesSidebarPath('/bank-soal-arsip', '/bank-soal')).toBe(false);
		expect(matchesSidebarPath('/asesmen-kegiatan', '/asesmen')).toBe(false);
	});

	it('selects only the most specific active href to avoid double highlights', () => {
		expect(findActiveSidebarHref('/bank-soal', items)).toBe('/bank-soal');
		expect(findActiveSidebarHref('/bank-soal/daftar', items)).toBe('/bank-soal/daftar');
		expect(findActiveSidebarHref('/bank-soal/daftar/abc', items)).toBe('/bank-soal/daftar');
		expect(findActiveSidebarHref('/bank-soal/analisis-butir', items)).toBe('/bank-soal/analisis-butir');
	});

	it('keeps nested assessment pages on their exact workflow item instead of parent dashboard', () => {
		expect(findActiveSidebarHref('/asesmen', items)).toBe('/asesmen');
		expect(findActiveSidebarHref('/asesmen/kegiatan', items)).toBe('/asesmen/kegiatan');
		expect(findActiveSidebarHref('/asesmen/kegiatan/new', items)).toBe('/asesmen/kegiatan/new');
		expect(findActiveSidebarHref('/asesmen/kegiatan/event-1/members', items)).toBe('/asesmen/kegiatan');
	});
});
