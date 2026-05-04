import { describe, expect, it } from 'vitest';
import { sidebarNavGroups } from './sidebar-config';

const cbtItems = sidebarNavGroups.find((group) => group.group === 'CBT')?.items ?? [];

describe('sidebar CBT configuration', () => {
	it('keeps CBT navigation aligned to the event-centered workflow', () => {
		expect(cbtItems.map((item) => item.label)).toEqual([
			'Kegiatan Ujian',
			'Bank Soal',
			'Review Soal',
			'Paket Ujian',
			'Sesi Ujian',
			'Ruang Saya',
			'Kesiapan BYOD',
			'Non-Tes'
		]);
	});

	it('does not expose retired or event-specific dead links', () => {
		expect(cbtItems.map((item) => item.href)).toEqual([
			'/cbt/events',
			'/cbt/soal',
			'/cbt/soal/review',
			'/cbt/packages',
			'/cbt/sessions',
			'/cbt/proctoring/rooms',
			'/cbt/byod',
			'/cbt/non-test'
		]);
		expect(cbtItems.some((item) => item.href === '/cbt/questions')).toBe(false);
		expect(cbtItems.some((item) => item.href.includes('[id]'))).toBe(false);
	});

	it('keeps operational setup admin-only and authoring/review visible to teacher roles', () => {
		expect(cbtItems.find((item) => item.href === '/cbt/events')?.roles).toEqual(['admin']);
		expect(cbtItems.find((item) => item.href === '/cbt/packages')?.roles).toEqual(['admin']);
		expect(cbtItems.find((item) => item.href === '/cbt/sessions')?.roles).toEqual(['admin']);
		expect(cbtItems.find((item) => item.href === '/cbt/soal')?.roles).toEqual(['admin', 'guru', 'reviewer']);
		expect(cbtItems.find((item) => item.href === '/cbt/soal/review')?.roles).toEqual(['admin', 'guru', 'reviewer']);
	});
});
