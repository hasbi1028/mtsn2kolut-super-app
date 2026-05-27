import { describe, expect, it } from 'vitest';
import { filterSidebarNavGroupsByAccess } from './sidebar-access';
import { dashboardNavItem, sidebarNavGroups } from './sidebar-config';
import sidebarIconSource from './SidebarIcon.svelte?raw';
import { flattenSidebarNavGroups, sidebarBreadcrumbLabel } from './sidebar-tree';

const flatItems = flattenSidebarNavGroups(sidebarNavGroups);
const hrefs = flatItems.map((item) => item.href);
const byHref = new Map(flatItems.map((item) => [item.href, item]));
const labelsByGroup = (group: string) => flatItems.filter((item) => item.group === group).map((item) => item.label);
const hrefsByGroup = (group: string) => flatItems.filter((item) => item.group === group).map((item) => item.href);

describe('sidebar 3-level full route coverage configuration', () => {
	it('keeps Dashboard only as the quick-access root item, not duplicated in a sidebar group', () => {
		expect(dashboardNavItem).toMatchObject({ href: '/', label: 'Dashboard', permissions: ['dashboard.read'], pinnable: false });
		expect(sidebarNavGroups.some((group) => group.group === 'Utama')).toBe(false);
		expect(hrefs).not.toContain('/');
	});

	it('organizes navigation into madrasah work areas without duplicate leaf hrefs', () => {
		expect(sidebarNavGroups.map((group) => group.group)).toEqual([
			'Beranda',
			'Portal',
			'Akademik',
			'Siswa & Orang Tua',
			'Nilai & Rapor',
			'Bank Soal',
			'Asesmen Ujian',
			'Tata Usaha',
			'Aset & Layanan',
			'Website',
			'Pegawai & Kehadiran',
			'Pengaturan'
		]);
		expect(new Set(hrefs).size).toBe(hrefs.length);
		expect(labelsByGroup('Nilai & Rapor')).toEqual(['Input Nilai', 'Rapor Siswa']);
	});

	it('supports 3-level breadcrumbs for nested sidebar leaves', () => {
		expect(sidebarBreadcrumbLabel(byHref.get('/bank-soal/tambah')!)).toBe('Bank Soal › Kelola Soal › Tambah Soal');
		expect(sidebarBreadcrumbLabel(byHref.get('/asesmen/pengawasan')!)).toBe('Asesmen Ujian › Alur Utama › Pantau Ruang');
		expect(sidebarBreadcrumbLabel(byHref.get('/settings/backups')!)).toBe('Pengaturan › Sistem & Audit › Backup & Restore');
	});

	it('adds route coverage for important admin index/action pages while excluding dynamic detail routes', () => {
		expect(hrefs).toEqual(expect.arrayContaining([
			'/bank-soal/tambah',
			'/bank-soal/impor',
			'/bank-soal/cetak',
			'/bank-soal/mapel-kd',
			'/bank-soal/laporan',
			'/asesmen/ringkas',
			'/asesmen/persiapan',
			'/asesmen/pelaksanaan',
			'/asesmen/pengawasan',
			'/asesmen/hasil',
			'/governance/actions/calendar',
			'/governance/actions/meeting-pack',
			'/settings/maintenance'
		]));
		expect(hrefs.some((href) => href.includes('[id]') || href.includes('[slug]') || href.includes('[token]'))).toBe(false);
		expect(hrefs.some((href) => href.startsWith('/cbt/questions'))).toBe(false);
	});

	it('keeps assessment navigation aligned to preparation, execution, and result workflows', () => {
		expect(hrefsByGroup('Asesmen Ujian')).toEqual([
			'/asesmen/ringkas',
			'/asesmen/persiapan',
			'/asesmen/pelaksanaan',
			'/asesmen/pengawasan',
			'/asesmen/hasil'
		]);
		expect(labelsByGroup('Asesmen Ujian')).toEqual([
			'Ringkasan Ujian',
			'Persiapan',
			'Pelaksanaan',
			'Pantau Ruang',
			'Hasil',
		]);
		expect(hrefsByGroup('Asesmen Ujian')).not.toEqual(expect.arrayContaining([
			'/asesmen/kegiatan/new',
			'/asesmen/paket/new',
			'/asesmen/sesi/new',
			'/asesmen/aplikasi-siswa/matrix',
			'/asesmen/aplikasi-siswa/release',
			'/asesmen',
			'/asesmen/non-tes',
			'/asesmen/aplikasi-siswa',
			'/ujian'
		]));
		expect(hrefsByGroup('Asesmen Ujian').some((href) => href.startsWith('/bank-soal'))).toBe(false);
	});

	it('separates Bank Soal as a standalone module outside CBT routes', () => {
		expect(sidebarNavGroups.findIndex((group) => group.group === 'Bank Soal')).toBeLessThan(
			sidebarNavGroups.findIndex((group) => group.group === 'Asesmen Ujian')
		);
		expect(sidebarNavGroups.some((group) => group.group === 'CBT')).toBe(false);
		expect(hrefsByGroup('Bank Soal')).toEqual([
			'/bank-soal',
			'/bank-soal/daftar',
			'/bank-soal/tambah',
			'/bank-soal/impor',
			'/bank-soal/cetak',
			'/bank-soal/verifikasi',
			'/bank-soal/penerbitan',
			'/bank-soal/mapel-kd',
			'/bank-soal/analisis-butir',
			'/bank-soal/laporan',
			'/bank-soal/pengaturan'
		]);
		expect(hrefsByGroup('Bank Soal').every((href) => !href.startsWith('/cbt/'))).toBe(true);
	});

	it('keeps every configured sidebar icon backed by a rendered SVG branch', () => {
		const configuredIcons = new Set([dashboardNavItem.icon, ...flatItems.map((item) => item.icon)]);
		for (const icon of configuredIcons) {
			expect(sidebarIconSource).toContain(`name === '${icon}'`);
		}
	});

	it('adds permission metadata for migrated RBAC-aware modules while keeping role fallback', () => {
		expect(flatItems.every((item) => item.permissions.length > 0)).toBe(true);
		expect(byHref.get('/settings/account')).toMatchObject({ permissions: ['settings.account'], allowAuthenticatedFallback: true });
		expect(byHref.get('/settings/users')?.permissions).toEqual(['users.read']);
		expect(byHref.get('/settings/rbac')?.permissions).toEqual(['roles.read']);
		expect(byHref.get('/bank-soal/verifikasi')?.permissions).toEqual(['bank_soal.review', 'bank_soal.publish']);
		expect(byHref.get('/bank-soal/analisis-butir')?.permissions).toEqual(['bank_soal.analytics']);
		expect(byHref.get('/asesmen/persiapan')?.permissions).toEqual(['asesmen.read']);
		expect(byHref.get('/asesmen/pelaksanaan')?.permissions).toEqual(['asesmen.proctor']);
		expect(byHref.get('/asesmen/pengawasan')?.permissions).toEqual(['asesmen.proctor']);
		expect(byHref.get('/asesmen')).toBeUndefined();
		expect(byHref.get('/asesmen/non-tes')).toBeUndefined();
		expect(byHref.get('/asesmen/aplikasi-siswa')).toBeUndefined();
		expect(byHref.get('/asesmen/aplikasi-siswa/matrix')).toBeUndefined();
		expect(byHref.get('/asesmen/aplikasi-siswa/release')).toBeUndefined();
	});

	it('hides guru role-only academic, student, assessment, and bank-soal surfaces without permissions', () => {
		const visibleHrefs = flattenSidebarNavGroups(filterSidebarNavGroupsByAccess(sidebarNavGroups, ['guru'], [])).map((item) => item.href);
		expect(visibleHrefs).toEqual(['/settings/account']);
		expect(visibleHrefs).not.toContain('/students');
		expect(visibleHrefs).not.toContain('/bank-soal');
		expect(visibleHrefs).not.toContain('/asesmen/pelaksanaan');
	});

	it('shows Bank Soal read/create surfaces for a guru with matching permissions', () => {
		const visibleHrefs = flattenSidebarNavGroups(
			filterSidebarNavGroupsByAccess(sidebarNavGroups, ['guru'], ['bank_soal.read', 'bank_soal.create'])
		).map((item) => item.href);
		expect(visibleHrefs).toEqual(expect.arrayContaining(['/bank-soal', '/bank-soal/daftar', '/bank-soal/tambah', '/bank-soal/mapel-kd', '/settings/account']));
		expect(visibleHrefs).not.toContain('/bank-soal/verifikasi');
	});

	it('exposes student and parent portal entries only to matching roles or permissions', () => {
		const visibleHrefs = (roles: string[], permissions: string[] = []) =>
			flattenSidebarNavGroups(filterSidebarNavGroupsByAccess(sidebarNavGroups, roles, permissions)).map((item) => item.href);

		expect(hrefsByGroup('Portal')).toEqual(['/portal/siswa', '/portal/siswa/qr-login', '/portal/orang-tua', '/jadwal']);
		expect(byHref.get('/portal/siswa')).toMatchObject({ roleFallbacks: ['siswa'], permissions: ['student_portal.read'] });
		expect(byHref.get('/portal/orang-tua')).toMatchObject({ roleFallbacks: ['ortu'], permissions: ['parent_portal.read'] });
		expect(visibleHrefs(['siswa'])).toContain('/portal/siswa');
		expect(visibleHrefs(['siswa'])).not.toContain('/portal/orang-tua');
		expect(visibleHrefs(['ortu'])).toContain('/portal/orang-tua');
		expect(visibleHrefs(['ortu'])).not.toContain('/portal/siswa');
		expect(visibleHrefs([], ['student_portal.read'])).toContain('/portal/siswa');
		expect(visibleHrefs([], ['parent_portal.read'])).toContain('/portal/orang-tua');
	});
});
