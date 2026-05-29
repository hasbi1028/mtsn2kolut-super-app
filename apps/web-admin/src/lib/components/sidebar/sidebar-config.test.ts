import { describe, expect, it } from 'vitest';
import { filterSidebarNavGroupsByAccess } from './sidebar-access';
import { dashboardNavItem, sidebarNavGroups } from './sidebar-config';
import sidebarIconSource from './SidebarIcon.svelte?raw';
import { flattenSidebarNavGroups, numberSidebarNavGroups, sidebarBreadcrumbLabel, sidebarNumberedBreadcrumbLabel } from './sidebar-tree';

const flatItems = flattenSidebarNavGroups(sidebarNavGroups);
const numberedGroups = numberSidebarNavGroups(sidebarNavGroups);
const numberedFlatItems = flattenSidebarNavGroups(numberedGroups);
const hrefs = flatItems.map((item) => item.href);
const byHref = new Map(flatItems.map((item) => [item.href, item]));
const numberedByHref = new Map(numberedFlatItems.map((item) => [item.href, item]));
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
			'Asesmen',
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
		expect(sidebarBreadcrumbLabel(byHref.get('/bank-soal/tambah')!)).toBe('Bank Soal › Tambah Soal');
		expect(sidebarBreadcrumbLabel(byHref.get('/asesmen')!)).toBe('Asesmen › Ringkasan Asesmen');
		expect(sidebarBreadcrumbLabel(byHref.get('/settings/backups')!)).toBe('Pengaturan › Sistem & Audit › Backup & Restore');
	});

	it('assigns stable global numbers from the full sidebar tree', () => {
		expect(numberedGroups.map((group) => `${group.section} ${group.group}`)).toEqual([
			'1 Beranda',
			'2 Portal',
			'3 Akademik',
			'4 Siswa & Orang Tua',
			'5 Nilai & Rapor',
			'6 Bank Soal',
			'7 Asesmen',
			'8 Tata Usaha',
			'9 Aset & Layanan',
			'10 Website',
			'11 Pegawai & Kehadiran',
			'12 Pengaturan'
		]);
		expect(numberedByHref.get('/bank-soal/tambah')).toMatchObject({ section: '6.2', numberedLabel: '6.2 Tambah Soal' });
		expect(numberedByHref.get('/asesmen/persiapan')).toMatchObject({ groupSection: '7', ancestorSections: [], section: '7.1' });
		expect(numberedByHref.get('/asesmen/ruang-saya')).toMatchObject({ groupSection: '7', ancestorSections: [], section: '7.2.2' });
		expect(sidebarNumberedBreadcrumbLabel(numberedByHref.get('/settings/backups')!)).toBe(
			'12 Pengaturan › 12.3 Sistem & Audit › 12.3.3 Backup & Restore'
		);
	});

	it('adds route coverage for important admin index/action pages while excluding dynamic detail routes', () => {
		expect(hrefs).toEqual(expect.arrayContaining([
			'/bank-soal/tambah',
			'/bank-soal/impor',
			'/asesmen',
			'/asesmen/persiapan',
			'/asesmen/pelaksanaan',
			'/asesmen/hasil',
			'/governance/actions/calendar',
			'/governance/actions/meeting-pack',
			'/settings/maintenance'
		]));
		expect(hrefs.some((href) => href.includes('[id]') || href.includes('[slug]') || href.includes('[token]'))).toBe(false);
		expect(hrefs.some((href) => href.startsWith('/cbt/questions'))).toBe(false);
	});

	it('keeps assessment navigation aligned to preparation, execution, and result workflows', () => {
		expect(hrefsByGroup('Asesmen')).toEqual([
			'/asesmen',
			'/asesmen/persiapan',
			'/asesmen/kegiatan',
			'/asesmen/paket',
			'/asesmen/sesi',
			'/asesmen/sesi#ruang-peserta',
			'/asesmen/sesi#pengawas',
			'/asesmen/pelaksanaan',
			'/asesmen/sesi?schedule=today',
			'/asesmen/ruang-saya',
			'/asesmen/ruang-saya#panel-ruang',
			'/asesmen/aplikasi-siswa',
			'/asesmen/sesi?area=serah-terima',
			'/asesmen/hasil',
			'/asesmen/hasil#rekap-nilai',
			'/asesmen/hasil#status-submit',
			'/asesmen/hasil#koreksi-uraian',
			'/asesmen/hasil#analisis-butir',
			'/asesmen/hasil#sinkronisasi',
			'/asesmen/kegiatan?arsip=utama',
			'/asesmen/sesi?dokumen=berita-acara',
			'/asesmen/hasil#rekap-pelaksanaan',
			'/asesmen/kegiatan?arsip=tindak-lanjut'
		]);
		expect(labelsByGroup('Asesmen')).toEqual([
			'Ringkasan Asesmen',
			'Persiapan',
			'Kegiatan Asesmen',
			'Paket Soal',
			'Sesi Ujian',
			'Ruang & Peserta',
			'Pengawas Ruang',
			'Pelaksanaan',
			'Sesi Panitia',
			'Ruang Saya',
			'Panel Ruang',
			'Perangkat Siswa',
			'Serah Terima Pengawas',
			'Hasil',
			'Rekap Nilai',
			'Status Submit',
			'Koreksi Uraian',
			'Analisis Butir',
			'Publikasi / Sinkronisasi',
			'Arsip',
			'Berita Acara',
			'Rekap Pelaksanaan',
			'Tindak Lanjut Sesi',
		]);
		expect(numberedByHref.get('/asesmen/kegiatan')).toMatchObject({ section: '7.1.1' });
		expect(numberedByHref.get('/asesmen/sesi?dokumen=berita-acara')).toMatchObject({ section: '7.4.1' });
		expect(hrefsByGroup('Asesmen')).not.toEqual(expect.arrayContaining([
			'/asesmen/kegiatan/new',
			'/asesmen/paket/new',
			'/asesmen/sesi/new',
			'/asesmen/aplikasi-siswa/matrix',
			'/asesmen/aplikasi-siswa/release',
			'/asesmen/non-tes',
			'/asesmen/aplikasi-siswa',
			'/asesmen/pengawasan',
			'/asesmen/panitia',
			'/ujian'
		]));
		expect(hrefsByGroup('Asesmen').some((href) => href.startsWith('/bank-soal'))).toBe(false);
	});

	it('separates Bank Soal as a standalone module outside CBT routes', () => {
		expect(sidebarNavGroups.findIndex((group) => group.group === 'Bank Soal')).toBeLessThan(
			sidebarNavGroups.findIndex((group) => group.group === 'Asesmen')
		);
		expect(sidebarNavGroups.some((group) => group.group === 'CBT')).toBe(false);
		expect(hrefsByGroup('Bank Soal')).toEqual([
			'/bank-soal',
			'/bank-soal/tambah',
			'/bank-soal/impor',
			'/bank-soal/verifikasi'
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
		expect(byHref.get('/bank-soal/analisis-butir')).toBeUndefined();
		expect(byHref.get('/asesmen')?.permissions).toEqual(['asesmen.proctor', 'asesmen.operator', 'asesmen.event_manage', 'asesmen.package_manage', 'asesmen.session_manage', 'asesmen.participant_manage', 'asesmen.result_read', 'asesmen.result_manage']);
		expect(byHref.get('/asesmen/ringkas')).toBeUndefined();
		expect(byHref.get('/asesmen/persiapan')?.permissions).toEqual(['asesmen.operator', 'asesmen.event_manage', 'asesmen.package_manage', 'asesmen.session_manage', 'asesmen.participant_manage']);
		expect(byHref.get('/asesmen/pelaksanaan')?.permissions).toEqual(['asesmen.proctor', 'asesmen.operator', 'asesmen.event_manage', 'asesmen.package_manage', 'asesmen.session_manage']);
		expect(byHref.get('/asesmen/hasil')?.permissions).toEqual(['asesmen.result_read', 'asesmen.result_manage']);
		expect(byHref.get('/asesmen/ruang-saya')?.section).toBe('7.2.2');
		expect(byHref.get('/asesmen/pengawasan')).toBeUndefined();
		expect(byHref.get('/asesmen/non-tes')).toBeUndefined();
		expect(byHref.get('/asesmen/aplikasi-siswa')?.section).toBe('7.2.4');
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


	it('shows result entry only for result readers/managers or admin fallback', () => {
		const readerHrefs = flattenSidebarNavGroups(
			filterSidebarNavGroupsByAccess(sidebarNavGroups, ['guru'], ['asesmen.result_read'])
		).map((item) => item.href);
		const managerHrefs = flattenSidebarNavGroups(
			filterSidebarNavGroupsByAccess(sidebarNavGroups, ['guru'], ['asesmen.result_manage'])
		).map((item) => item.href);
		expect(readerHrefs).toEqual(expect.arrayContaining(['/asesmen', '/asesmen/hasil', '/settings/account']));
		expect(managerHrefs).toEqual(expect.arrayContaining(['/asesmen', '/asesmen/hasil', '/settings/account']));
		expect(readerHrefs).not.toContain('/asesmen/persiapan');
		expect(managerHrefs).not.toContain('/asesmen/persiapan');
	});

	it('shows Bank Soal read/create surfaces for a guru with matching permissions', () => {
		const visibleHrefs = flattenSidebarNavGroups(
			filterSidebarNavGroupsByAccess(sidebarNavGroups, ['guru'], ['bank_soal.read', 'bank_soal.create'])
		).map((item) => item.href);
		expect(visibleHrefs).toEqual(expect.arrayContaining(['/bank-soal', '/bank-soal/tambah', '/settings/account']));
		expect(visibleHrefs).not.toContain('/bank-soal/verifikasi');
		expect(visibleHrefs).not.toContain('/bank-soal/daftar');
	});

	it('shows only the field-facing assessment entry for a proctor permission set', () => {
		const visibleHrefs = flattenSidebarNavGroups(
			filterSidebarNavGroupsByAccess(sidebarNavGroups, ['guru'], ['asesmen.proctor'])
		).map((item) => item.href);
		expect(visibleHrefs).toEqual(expect.arrayContaining(['/asesmen', '/asesmen/pelaksanaan', '/asesmen/ruang-saya', '/settings/account']));
		expect(visibleHrefs).not.toContain('/asesmen/ringkas');
		expect(visibleHrefs).not.toContain('/asesmen/persiapan');
	});

	it('shows the operator assessment control lane with stable work-code aliases', () => {
		const visibleHrefs = flattenSidebarNavGroups(
			filterSidebarNavGroupsByAccess(sidebarNavGroups, ['guru'], ['asesmen.operator'])
		).map((item) => item.href);
		expect(visibleHrefs).toEqual(expect.arrayContaining(['/asesmen', '/asesmen/persiapan', '/asesmen/pelaksanaan', '/asesmen/ruang-saya', '/settings/account']));
		expect(visibleHrefs).not.toContain('/asesmen/ringkas');
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
