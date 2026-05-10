import { describe, expect, it } from 'vitest';
import { filterSidebarNavGroupsByAccess } from './sidebar-access';
import { dashboardNavItem, sidebarNavGroups } from './sidebar-config';
import sidebarIconSource from './SidebarIcon.svelte?raw';

const academicItems = sidebarNavGroups.find((group) => group.group === 'Akademik & Pembelajaran')?.items ?? [];
const assessmentItems = sidebarNavGroups.find((group) => group.group === 'Asesmen')?.items ?? [];
const bankSoalItems = sidebarNavGroups.find((group) => group.group === 'Bank Soal')?.items ?? [];
const portalItems = sidebarNavGroups.find((group) => group.group === 'Portal')?.items ?? [];

describe('sidebar assessment configuration', () => {
	it('keeps Dashboard only as the quick-access root item, not duplicated in a Utama group', () => {
		expect(dashboardNavItem).toMatchObject({ href: '/', label: 'Dashboard', permissions: ['dashboard.read'], pinnable: false });
		expect(sidebarNavGroups.some((group) => group.group === 'Utama')).toBe(false);
		expect(sidebarNavGroups.flatMap((group) => group.items).some((item) => item.href === '/')).toBe(false);
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
			'Dashboard Asesmen',
			'Paket Soal',
			'Kegiatan',
			'Persiapan',
			'APK CBT Mobile',
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
			'/asesmen/aplikasi-siswa/release',
			'/asesmen/pelaksanaan',
			'/asesmen/hasil'
		]);
		expect(assessmentItems).toHaveLength(7);
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
		expect(assessmentItems.find((item) => item.href === '/asesmen/aplikasi-siswa/release')?.roles).toEqual(['admin', 'guru', 'staf']);
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
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/verifikasi')?.roles).toEqual(['admin']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/impor')?.roles).toEqual(['admin']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/pengaturan')?.roles).toEqual(['admin']);
		expect(bankSoalItems.filter((item) => !['/bank-soal/verifikasi', '/bank-soal/impor', '/bank-soal/pengaturan'].includes(item.href)).every((item) => item.roles?.includes('admin') && item.roles.includes('guru'))).toBe(true);
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
			label: 'Manajemen RBAC',
			roles: ['admin'],
			permissions: ['roles.read']
		});
		expect(byHref.get('/settings/user-change-requests')?.permissions).toEqual(['profile_changes.review']);
		expect(byHref.get('/settings/audit-logs')?.permissions).toEqual(['audit.read']);
		expect(byHref.get('/settings/analytics')).toMatchObject({
			label: 'Analytics Internal',
			roles: ['admin'],
			permissions: ['analytics.read']
		});
		expect(byHref.get('/settings/school-profile')?.permissions).toEqual(['settings.school_profile']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/tambah')?.permissions).toEqual(['bank_soal.create']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/verifikasi')?.permissions).toEqual(['bank_soal.review']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/impor')?.permissions).toEqual(['bank_soal.import']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/analisis-butir')?.permissions).toEqual(['bank_soal.analytics']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/pengaturan')?.permissions).toEqual(['bank_soal.settings']);
		expect(bankSoalItems.filter((item) => !['/bank-soal/tambah', '/bank-soal/verifikasi', '/bank-soal/impor', '/bank-soal/analisis-butir', '/bank-soal/pengaturan'].includes(item.href)).every((item) => item.permissions.includes('bank_soal.read'))).toBe(true);
		expect(assessmentItems.find((item) => item.href === '/asesmen/kegiatan')?.permissions).toEqual(['asesmen.event_manage']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/paket')?.permissions).toEqual(['asesmen.package_manage']);
		expect(assessmentItems.find((item) => item.href === '/asesmen/aplikasi-siswa/release')?.permissions).toEqual(['asesmen.read', 'asesmen.proctor']);
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
			'/bank-soal/tambah',
			'/bank-soal/mapel-kd',
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
