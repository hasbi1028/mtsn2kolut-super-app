import { describe, expect, it } from 'vitest';
import { sidebarNavGroups } from './sidebar-config';

const cbtItems = sidebarNavGroups.find((group) => group.group === 'CBT')?.items ?? [];

describe('sidebar CBT configuration', () => {
	it('keeps CBT navigation aligned to the event-centered workflow', () => {
		expect(cbtItems.map((item) => item.label)).toEqual([
			'Dashboard CBT',
			'Bank Soal',
			'Paket Soal',
			'Kegiatan & Sesi',
			'Monitoring',
			'Hasil & Analisis'
		]);
	});

	it('does not expose retired or event-specific dead links', () => {
		expect(cbtItems.map((item) => item.href)).toEqual([
			'/cbt',
			'/cbt/soal',
			'/cbt/packages',
			'/cbt/events',
			'/cbt/byod',
			'/cbt/hasil'
		]);
		expect(cbtItems).toHaveLength(6);
		expect(new Set(cbtItems.map((item) => item.href)).size).toBe(cbtItems.length);
		expect(cbtItems.some((item) => item.href === '/cbt/questions')).toBe(false);
		expect(cbtItems.some((item) => item.href.includes('[id]'))).toBe(false);
	});

	it('keeps operational setup admin-only and authoring/review visible to teacher roles', () => {
		expect(cbtItems.find((item) => item.href === '/cbt')?.roles).toEqual(['admin', 'guru', 'staf']);
		expect(cbtItems.find((item) => item.label === 'Kegiatan & Sesi')?.roles).toEqual(['admin']);
		expect(cbtItems.find((item) => item.href === '/cbt/packages')?.roles).toEqual(['admin']);
		expect(cbtItems.find((item) => item.href === '/cbt/soal')?.roles).toEqual(['admin', 'guru']);
		expect(cbtItems.find((item) => item.href === '/cbt/byod')?.roles).toEqual(['admin', 'guru', 'staf']);
		expect(cbtItems.find((item) => item.label === 'Hasil & Analisis')?.roles).toEqual(['admin', 'guru']);
		expect(cbtItems.flatMap((item) => item.roles ?? [])).not.toContain('reviewer');
	});
});
