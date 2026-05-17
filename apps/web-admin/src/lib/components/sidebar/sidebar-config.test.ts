import { describe, expect, it } from 'vitest';
import { filterSidebarNavGroupsByAccess } from './sidebar-access';
import { dashboardNavItem, sidebarNavGroups } from './sidebar-config';
import sidebarIconSource from './SidebarIcon.svelte?raw';

const academicItems = sidebarNavGroups.find((group) => group.group === 'Akademik')?.items ?? [];
const assessmentItems = sidebarNavGroups.find((group) => group.group === 'Asesmen / CBT')?.items ?? [];
const bankSoalItems = sidebarNavGroups.find((group) => group.group === 'Bank Soal')?.items ?? [];
const portalItems = sidebarNavGroups.find((group) => group.group === 'Portal')?.items ?? [];

describe('sidebar assessment configuration', () => {
	it('keeps Dashboard only as the quick-access root item, not duplicated in a Utama group', () => {
		expect(dashboardNavItem).toMatchObject({ href: '/', label: 'Dashboard', permissions: ['dashboard.read'], pinnable: false });
		expect(sidebarNavGroups.some((group) => group.group === 'Utama')).toBe(false);
		expect(sidebarNavGroups.flatMap((group) => group.items).some((item) => item.href === '/')).toBe(false);
	});

	it('organizes navigation into madrasah work areas without duplicate hrefs', () => {
		expect(sidebarNavGroups.map((group) => group.group)).toEqual([
			'Beranda',
			'Portal',
			'Akademik',
			'Siswa & Orang Tua',
			'Nilai & Rapor',
			'Bank Soal',
			'Asesmen / CBT',
			'Tata Usaha',
			'Aset & Layanan',
			'Website',
			'Pegawai & Kehadiran',
			'Pengaturan'
		]);
		const hrefs = sidebarNavGroups.flatMap((group) => group.items.map((item) => item.href));
		expect(new Set(hrefs).size).toBe(hrefs.length);
		expect(sidebarNavGroups.find((group) => group.group === 'Nilai & Rapor')?.items.map((item) => item.label)).toEqual(['Input Nilai', 'Rapor Siswa']);
	});

	it('exposes Rombel under Akademik with RBAC fallback and permission metadata', () => {
		const rombel = academicItems.find((item) => item.href === '/akademik/rombel');
		expect(rombel).toMatchObject({
			label: 'Rombel',
			icon: 'layers',
			roles: ['admin', 'guru', 'kesiswaan'],
			permissions: ['academic.read']
		});
		expect(academicItems.filter((item) => item.href === '/akademik/rombel')).toHaveLength(1);
	});

	it('keeps assessment navigation aligned to the three-phase workflow', () => {
		expect(assessmentItems.map((item) => item.label)).toEqual([
			'Hari Ini / Dashboard CBT',
			'Persiapan Ujian',
			'Monitor Ujian',
			'Hasil & BA',
			'Arsip',
			'Aplikasi Siswa'
		]);
	});

	it('does not expose Bank Soal authoring links in the assessment group', () => {
		expect(assessmentItems.map((item) => item.href)).toEqual([
			'/asesmen',
			'/asesmen/persiapan',
			'/asesmen/pelaksanaan',
			'/asesmen/hasil',
			'/asesmen/kegiatan',
			'/asesmen/aplikasi-siswa'
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
		expect(assessmentItems.find((item) => item.href === '/asesmen/persiapan')?.roles).toEqual(['admin', 'guru']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/pelaksanaan')?.roles).toEqual(['admin', 'guru', 'staf']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/hasil')?.roles).toEqual(['admin', 'guru']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/kegiatan')?.roles).toEqual(['admin', 'guru']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/aplikasi-siswa')?.roles).toEqual(['admin', 'guru', 'staf']);
		expect(assessmentItems.flatMap((item) => item.roles ?? [])).not.toContain('reviewer');
	});

	it('separates Bank Soal as a standalone module outside CBT routes', () => {
		expect(sidebarNavGroups.findIndex((group) => group.group === 'Bank Soal')).toBeLessThan(
			sidebarNavGroups.findIndex((group) => group.group === 'Asesmen / CBT')
		);
		expect(sidebarNavGroups.some((group) => group.group === 'CBT')).toBe(false);
		expect(sidebarNavGroups.some((group) => group.group === 'Bank Soal & Asesmen')).toBe(false);
		expect(bankSoalItems.map((item) => item.label)).toEqual([
			'Dashboard Bank Soal',
			'Kelola Soal',
			'Review & Terbitkan',
			'Mutu Soal',
			'Pengaturan'
		]);
		expect(bankSoalItems.map((item) => item.href)).toEqual([
			'/bank-soal',
			'/bank-soal/daftar',
			'/bank-soal/verifikasi',
			'/bank-soal/analisis-butir',
			'/bank-soal/pengaturan'
		]);
		expect(bankSoalItems.every((item) => !item.href.includes('?mode='))).toBe(true);
		expect(bankSoalItems.map((item) => item.icon)).toEqual([
			'grid',
			'book-open',
			'clipboard',
			'activity',
			'settings'
		]);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/verifikasi')?.roles).toEqual(['admin']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/pengaturan')?.roles).toEqual(['admin']);
		expect(bankSoalItems.filter((item) => !['/bank-soal/verifikasi', '/bank-soal/pengaturan'].includes(item.href)).every((item) => item.roles?.includes('admin') && item.roles.includes('guru'))).toBe(true);
		expect(bankSoalItems.every((item) => !item.href.startsWith('/cbt/'))).toBe(true);
		expect(bankSoalItems.some((item) => item.href === '/cbt/questions')).toBe(false);
	});



	it('keeps every configured sidebar icon backed by a rendered SVG branch', () => {
		const configuredIcons = new Set([
			dashboardNavItem.icon,
			...sidebarNavGroups.flatMap((group) => group.items.map((item) => item.icon))
		]);

		for (const icon of configuredIcons) {
			expect(sidebarIconSource).toContain(`name === '${icon}'`);
		}
	});

	it('adds permission metadata for migrated RBAC-aware modules while keeping role fallback', () => {
		const allItems = sidebarNavGroups.flatMap((group) => group.items);
		const byHref = new Map(allItems.map((item) => [item.href, item]));

		expect(allItems.every((item) => item.permissions.length > 0)).toBe(true);
		expect(byHref.get('/settings/account')).toMatchObject({
			permissions: ['settings.account'],
			allowAuthenticatedFallback: true
		});
		expect(byHref.get('/settings')?.permissions).toEqual(['settings.account']);
		expect(byHref.get('/settings/users')?.permissions).toEqual(['users.read']);
		expect(byHref.get('/settings/rbac')).toMatchObject({
			label: 'Peran & Izin Akses',
			roles: ['admin'],
			permissions: ['roles.read']
		});
		expect(byHref.get('/settings/user-change-requests')?.permissions).toEqual(['profile_changes.review']);
		expect(byHref.get('/settings/audit-logs')?.permissions).toEqual(['audit.read']);
		expect(byHref.get('/settings/analytics')).toMatchObject({
			label: 'Statistik Penggunaan',
			roles: ['admin'],
			permissions: ['analytics.read']
		});
		expect(byHref.get('/settings/school-profile')?.permissions).toEqual(['settings.school_profile']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/verifikasi')?.permissions).toEqual(['bank_soal.review', 'bank_soal.publish']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/analisis-butir')?.permissions).toEqual(['bank_soal.analytics']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/pengaturan')?.permissions).toEqual(['bank_soal.settings']);
		expect(bankSoalItems.filter((item) => !['/bank-soal/verifikasi', '/bank-soal/analisis-butir', '/bank-soal/pengaturan'].includes(item.href)).every((item) => item.permissions.includes('bank_soal.read'))).toBe(true);
		expect(assessmentItems.find((item) => item.href === '/asesmen/kegiatan')?.permissions).toEqual(['asesmen.read', 'asesmen.event_manage']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/aplikasi-siswa')?.permissions).toEqual(['asesmen.read', 'asesmen.proctor']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/hasil')?.permissions).toEqual(['asesmen.result_read']);
	});

	it('hides guru role-only academic, student, assessment, and bank-soal surfaces without permissions', () => {
		const visibleHrefs = filterSidebarNavGroupsByAccess(sidebarNavGroups, ['guru'], [])
			.flatMap((group) => group.items.map((item) => item.href));

		expect(visibleHrefs).toEqual(['/settings/account']);
		expect(visibleHrefs).not.toContain('/jadwal');
		expect(visibleHrefs).not.toContain('/students');
		expect(visibleHrefs).not.toContain('/bank-soal');
		expect(visibleHrefs).not.toContain('/bank-soal/tambah');
		expect(visibleHrefs).not.toContain('/asesmen/pelaksanaan');
		expect(visibleHrefs).not.toContain('/asesmen/hasil');
	});

	it('shows only Bank Soal and account surfaces for a guru with Bank Soal read/create permissions', () => {
		const visibleHrefs = filterSidebarNavGroupsByAccess(sidebarNavGroups, ['guru'], ['bank_soal.read', 'bank_soal.create'])
			.flatMap((group) => group.items.map((item) => item.href));

		expect(visibleHrefs).toEqual([
			'/bank-soal',
			'/bank-soal/daftar',
			'/settings/account'
		]);
	});

	it('exposes student and parent portal entries only to matching roles or permissions', () => {
		const allItems = sidebarNavGroups.flatMap((group) => group.items);
		const byHref = new Map(allItems.map((item) => [item.href, item]));
		const visibleHrefs = (roles: string[], permissions: string[] = []) =>
			filterSidebarNavGroupsByAccess(sidebarNavGroups, roles, permissions)
				.flatMap((group) => group.items.map((item) => item.href));

		expect(portalItems.map((item) => item.href)).toEqual(['/portal/siswa', '/portal/orang-tua']);
		expect(byHref.get('/portal/siswa')).toMatchObject({
			label: 'Portal Siswa',
			icon: 'book-open',
			roles: ['siswa'],
			roleFallbacks: ['siswa'],
			permissions: ['student_portal.read']
		});
		expect(byHref.get('/portal/orang-tua')).toMatchObject({
			label: 'Portal Orang Tua',
			icon: 'user-group',
			roles: ['ortu'],
			roleFallbacks: ['ortu'],
			permissions: ['parent_portal.read']
		});
		expect(visibleHrefs(['siswa'])).toContain('/portal/siswa');
		expect(visibleHrefs(['siswa'])).not.toContain('/portal/orang-tua');
		expect(visibleHrefs(['ortu'])).toContain('/portal/orang-tua');
		expect(visibleHrefs(['ortu'])).not.toContain('/portal/siswa');
		expect(visibleHrefs([], ['student_portal.read'])).toContain('/portal/siswa');
		expect(visibleHrefs([], ['parent_portal.read'])).toContain('/portal/orang-tua');
	});

	it('does not expose the retired question-bank route anywhere in sidebar nav', () => {
		const allItems = sidebarNavGroups.flatMap((group) => group.items);
		expect(allItems.some((item) => item.href.startsWith('/cbt/questions'))).toBe(false);
	});
});
