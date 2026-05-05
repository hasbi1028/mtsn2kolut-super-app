import { describe, expect, it } from 'vitest';
import { sidebarNavGroups } from './sidebar-config';

const cbtItems = sidebarNavGroups.find((group) => group.group === 'CBT')?.items ?? [];

describe('sidebar CBT configuration', () => {
	it('keeps CBT navigation aligned to the three-phase workflow', () => {
		expect(cbtItems.map((item) => item.label)).toEqual([
			'Dashboard CBT',
			'Persiapan',
			'Pelaksanaan',
			'Hasil'
		]);
	});

	it('does not expose granular CBT task links in the sidebar', () => {
		expect(cbtItems.map((item) => item.href)).toEqual([
			'/cbt',
			'/cbt/persiapan',
			'/cbt/pelaksanaan',
			'/cbt/hasil'
		]);
		expect(cbtItems).toHaveLength(4);
		expect(new Set(cbtItems.map((item) => item.href)).size).toBe(cbtItems.length);
		expect(cbtItems.some((item) => item.href === '/cbt/soal')).toBe(false);
		expect(cbtItems.some((item) => item.href === '/cbt/bank-soal')).toBe(false);
		expect(cbtItems.some((item) => item.href === '/cbt/questions')).toBe(false);
		expect(cbtItems.some((item) => item.href.includes('[id]'))).toBe(false);
	});

	it('keeps CBT phase visibility aligned to role scope', () => {
		expect(cbtItems.find((item) => item.href === '/cbt')?.roles).toEqual(['admin', 'guru', 'staf']);
		expect(cbtItems.find((item) => item.href === '/cbt/persiapan')?.roles).toEqual(['admin', 'guru']);
		expect(cbtItems.find((item) => item.href === '/cbt/pelaksanaan')?.roles).toEqual(['admin', 'guru', 'staf']);
		expect(cbtItems.find((item) => item.href === '/cbt/hasil')?.roles).toEqual(['admin', 'guru']);
		expect(cbtItems.flatMap((item) => item.roles ?? [])).not.toContain('reviewer');
	});
});
