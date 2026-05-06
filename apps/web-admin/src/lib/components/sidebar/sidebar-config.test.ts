import { describe, expect, it } from 'vitest';
import { sidebarNavGroups } from './sidebar-config';

const assessmentItems = sidebarNavGroups.find((group) => group.group === 'Asesmen')?.items ?? [];
const bankSoalItems = sidebarNavGroups.find((group) => group.group === 'Bank Soal')?.items ?? [];

describe('sidebar assessment configuration', () => {
	it('keeps assessment navigation aligned to the three-phase workflow', () => {
		expect(assessmentItems.map((item) => item.label)).toEqual([
			'Dashboard Asesmen',
			'Paket Soal',
			'Kegiatan',
			'Persiapan',
			'Pelaksanaan',
			'Hasil'
		]);
	});

	it('does not expose Bank Soal authoring links in the assessment group', () => {
		expect(assessmentItems.map((item) => item.href)).toEqual([
			'/asesmen',
			'/asesmen/paket',
			'/asesmen/kegiatan',
			'/asesmen/persiapan',
			'/asesmen/pelaksanaan',
			'/asesmen/hasil'
		]);
		expect(assessmentItems).toHaveLength(6);
		expect(new Set(assessmentItems.map((item) => item.href)).size).toBe(assessmentItems.length);
		expect(assessmentItems.some((item) => item.href === '/cbt/soal')).toBe(false);
		expect(assessmentItems.some((item) => item.href === '/cbt/bank-soal')).toBe(false);
		expect(assessmentItems.some((item) => item.href === '/cbt/questions')).toBe(false);
		expect(assessmentItems.some((item) => item.href.startsWith('/bank-soal'))).toBe(false);
		expect(assessmentItems.some((item) => item.href.includes('[id]'))).toBe(false);
	});

	it('keeps assessment phase visibility aligned to role scope', () => {
		expect(assessmentItems.find((item) => item.href === '/asesmen')?.roles).toEqual(['admin', 'guru', 'staf']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/paket')?.roles).toEqual(['admin']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/kegiatan')?.roles).toEqual(['admin']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/persiapan')?.roles).toEqual(['admin', 'guru']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/pelaksanaan')?.roles).toEqual(['admin', 'guru', 'staf']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/hasil')?.roles).toEqual(['admin', 'guru']);
		expect(assessmentItems.flatMap((item) => item.roles ?? [])).not.toContain('reviewer');
	});

	it('separates Bank Soal as a standalone module outside CBT routes', () => {
		expect(sidebarNavGroups.findIndex((group) => group.group === 'Bank Soal')).toBeLessThan(
			sidebarNavGroups.findIndex((group) => group.group === 'Asesmen')
		);
		expect(sidebarNavGroups.some((group) => group.group === 'CBT')).toBe(false);
		expect(sidebarNavGroups.some((group) => group.group === 'Bank Soal & Asesmen')).toBe(false);
		expect(bankSoalItems.map((item) => item.label)).toEqual([
			'Dashboard Bank Soal',
			'Daftar Soal',
			'Tambah Soal',
			'Review Soal',
			'Impor Soal',
			'Analisis Butir',
			'Mapel & KD',
			'Pengaturan Bank Soal'
		]);
		expect(bankSoalItems.map((item) => item.href)).toEqual([
			'/bank-soal',
			'/bank-soal/daftar',
			'/bank-soal/tambah',
			'/bank-soal/verifikasi',
			'/bank-soal/impor',
			'/bank-soal/analisis-butir',
			'/bank-soal/mapel-kd',
			'/bank-soal/pengaturan'
		]);
		expect(bankSoalItems.every((item) => !item.href.includes('?mode='))).toBe(true);
		expect(bankSoalItems.map((item) => item.icon)).toEqual([
			'grid',
			'book-open',
			'pen-tool',
			'clipboard',
			'file-text',
			'activity',
			'layers',
			'settings'
		]);
		expect(bankSoalItems.every((item) => item.roles?.includes('admin') && item.roles.includes('guru'))).toBe(true);
		expect(bankSoalItems.every((item) => !item.href.startsWith('/cbt/'))).toBe(true);
		expect(bankSoalItems.some((item) => item.href === '/cbt/questions')).toBe(false);
	});



	it('adds permission metadata for migrated RBAC-aware modules while keeping role fallback', () => {
		const allItems = sidebarNavGroups.flatMap((group) => group.items);
		const byHref = new Map(allItems.map((item) => [item.href, item]));

		expect(byHref.get('/settings/users')?.permissions).toEqual(['users.read']);
		expect(byHref.get('/settings/audit-logs')?.permissions).toEqual(['audit.read']);
		expect(byHref.get('/settings/school-profile')?.permissions).toEqual(['settings.school_profile']);
		expect(bankSoalItems.every((item) => item.permissions?.includes('bank_soal.read'))).toBe(true);
		expect(assessmentItems.find((item) => item.href === '/asesmen/kegiatan')?.permissions).toEqual(['asesmen.event_manage']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/paket')?.permissions).toEqual(['asesmen.package_manage']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/hasil')?.permissions).toEqual(['asesmen.result_read']);
		expect(allItems.filter((item) => item.permissions?.length).every((item) => item.roles?.length || item.href === '/settings')).toBe(true);
	});

	it('does not expose the retired question-bank route anywhere in sidebar nav', () => {
		const allItems = sidebarNavGroups.flatMap((group) => group.items);
		expect(allItems.some((item) => item.href.startsWith('/cbt/questions'))).toBe(false);
	});
});
