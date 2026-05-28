import { describe, expect, it } from 'vitest';
import { buildAdminBreadcrumbs, compactAdminBreadcrumbs } from './admin-breadcrumb';

const adminRoles = ['admin'];
const adminPermissions = [
	'dashboard.read',
	'asesmen.operator',
	'asesmen.event_manage',
	'asesmen.package_manage',
	'asesmen.session_manage',
	'asesmen.participant_manage',
	'asesmen.result_read',
	'bank_soal.read',
	'bank_soal.create',
	'students.read',
	'academic.read',
	'letters.read',
	'letters.manage',
	'website.read',
	'website.manage',
	'library.read',
	'library.manage',
	'users.read',
	'pusaka.read'
];

const routedAdminModuleSamples = [
	'/notifications',
	'/akademik',
	'/academic',
	'/akademik/tahun-ajaran',
	'/akademik/kurikulum',
	'/akademik/mapel',
	'/akademik/rombel/01000000-0000-0000-0000-000000000000',
	'/akademik/guru-mapel',
	'/akademik/beban-guru',
	'/akademik/jam-pelajaran',
	'/akademik/jadwal',
	'/journal/01000000-0000-0000-0000-000000000000',
	'/students/01000000-0000-0000-0000-000000000000/edit',
	'/parents',
	'/kesiswaan',
	'/kesiswaan/kartu-siswa',
	'/kesiswaan/kartu-siswa/scan',
	'/grades',
	'/grades/rapor',
	'/bank-soal/soal/01000000-0000-0000-0000-000000000000',
	'/bank-soal/impor',
	'/bank-soal/verifikasi',
	'/asesmen/sesi/01000000-0000-0000-0000-000000000000/rooms/01000000-0000-0000-0000-000000000001/print-pack',
	'/tu',
	'/tu/surat-masuk',
	'/tu/surat-keluar',
	'/tu/surat-keterangan/42/print',
	'/tu/disposisi',
	'/tu/arsip',
	'/tu/compliance-pack',
	'/document-cycles',
	'/document-cycles/verifikasi',
	'/governance/actions/meeting-pack',
	'/governance/print-pack',
	'/library',
	'/library/books/abc123/copies',
	'/library/loans',
	'/inventory',
	'/inventory/items',
	'/website',
	'/website/posts/berita-madrasah/edit',
	'/website/announcements',
	'/website/pages',
	'/employees',
	'/pusaka/telegram-laporan',
	'/pusaka/antrian',
	'/settings/account',
	'/settings/users',
	'/settings/rbac',
	'/settings/backups',
	'/settings/audit-logs'
] as const;

describe('admin breadcrumb resolver', () => {
	it('uses sidebar trail for configured admin routes', () => {
		const asesmenCrumbs = buildAdminBreadcrumbs('/asesmen', adminRoles, adminPermissions);
		expect(asesmenCrumbs.map((crumb) => crumb.label)).toEqual([
			'Asesmen Ujian',
			'Alur Utama',
			'Ringkasan Asesmen'
		]);
		expect(asesmenCrumbs.map((crumb) => crumb.section)).toEqual(['7', '7.1', '7.1.1']);

		const tambahSoalCrumbs = buildAdminBreadcrumbs('/bank-soal/tambah', adminRoles, adminPermissions);
		expect(tambahSoalCrumbs.map((crumb) => crumb.label)).toEqual([
			'Bank Soal',
			'Kelola Soal',
			'Tambah Soal'
		]);
		expect(tambahSoalCrumbs.map((crumb) => crumb.section)).toEqual(['6', '6.1', '6.1.2']);
	});

	it('uses assessment fallback trails for routes hidden from sidebar', () => {
		expect(buildAdminBreadcrumbs('/asesmen/sesi/01000000-0000-0000-0000-000000000000', adminRoles, adminPermissions)).toEqual([
			{ label: 'Asesmen Ujian', href: undefined, section: '7' },
			{ label: 'Persiapan', href: '/asesmen/persiapan', section: '7.1.2' },
			{ label: 'Sesi & Ruang', href: '/asesmen/sesi', section: '7.1.2.3' },
			{ label: 'Detail Sesi', href: undefined, section: '7.1.2.3.1' }
		]);
		expect(buildAdminBreadcrumbs('/asesmen/ruang-saya', adminRoles, adminPermissions).map((crumb) => crumb.label)).toEqual([
			'Asesmen Ujian',
			'Pelaksanaan',
			'Ruang Saya'
		]);
		expect(buildAdminBreadcrumbs('/asesmen/ruang-saya', adminRoles, adminPermissions).map((crumb) => crumb.section)).toEqual([
			'7',
			'7.1.3',
			'7.1.3.1'
		]);
	});

	it('falls back to readable route segments for unknown admin routes', () => {
		expect(buildAdminBreadcrumbs('/unknown-tools/sync-log', adminRoles, adminPermissions).map((crumb) => crumb.label)).toEqual([
			'Dashboard',
			'Unknown Tools',
			'Sync Log'
		]);
	});

	it('extends numbered breadcrumbs for detail pages across modules', () => {
		expect(
			buildAdminBreadcrumbs('/students/01000000-0000-0000-0000-000000000000/edit', adminRoles, adminPermissions)
		).toEqual([
			{ label: 'Siswa & Orang Tua', href: undefined, section: '4' },
			{ label: 'Data Siswa', href: undefined, section: '4.1' },
			{ label: 'Siswa', href: '/students', section: '4.1.1' },
			{ label: 'Detail', href: undefined, section: '4.1.1.1' },
			{ label: 'Ubah', href: undefined, section: '4.1.1.1.1' }
		]);

		expect(buildAdminBreadcrumbs('/library/books/abc123/copies', adminRoles, adminPermissions)).toEqual([
			{ label: 'Aset & Layanan', href: undefined, section: '9' },
			{ label: 'Perpustakaan', href: undefined, section: '9.1' },
			{ label: 'Katalog Buku', href: '/library/books', section: '9.1.2' },
			{ label: 'Abc123', href: undefined, section: '9.1.2.1' },
			{ label: 'Eksemplar', href: undefined, section: '9.1.2.1.1' }
		]);
	});

	it('keeps numbered context for other major modules', () => {
		expect(buildAdminBreadcrumbs('/akademik/rombel/01000000-0000-0000-0000-000000000000', adminRoles, adminPermissions)).toEqual([
			{ label: 'Akademik', href: undefined, section: '3' },
			{ label: 'Rombel & Pengajaran', href: undefined, section: '3.2' },
			{ label: 'Rombel', href: '/akademik/rombel', section: '3.2.1' },
			{ label: 'Detail', href: undefined, section: '3.2.1.1' }
		]);

		expect(buildAdminBreadcrumbs('/tu/surat-keterangan/42/print', adminRoles, adminPermissions)).toEqual([
			{ label: 'Tata Usaha', href: undefined, section: '8' },
			{ label: 'Persuratan', href: undefined, section: '8.1' },
			{ label: 'Surat Keterangan', href: '/tu/surat-keterangan', section: '8.1.4' },
			{ label: 'Detail', href: undefined, section: '8.1.4.1' },
			{ label: 'Cetak', href: undefined, section: '8.1.4.1.1' }
		]);

		expect(buildAdminBreadcrumbs('/website/posts/berita-madrasah/edit', adminRoles, adminPermissions)).toEqual([
			{ label: 'Website', href: undefined, section: '10' },
			{ label: 'Kelola Konten', href: undefined, section: '10.1' },
			{ label: 'Kelola Berita', href: '/website/posts', section: '10.1.7' },
			{ label: 'Berita Madrasah', href: undefined, section: '10.1.7.1' },
			{ label: 'Ubah', href: undefined, section: '10.1.7.1.1' }
		]);
	});

	it('uses global sidebar numbering even when the current role cannot see that menu item', () => {
		expect(buildAdminBreadcrumbs('/pusaka/summary', ['guru'], [])).toEqual([
			{ label: 'Pegawai & Kehadiran', href: undefined, section: '11' },
			{ label: 'PUSAKA', href: undefined, section: '11.2' },
			{ label: 'Ringkasan Kehadiran', href: undefined, section: '11.2.4' }
		]);
	});

	it('keeps known admin module routes out of the generic dashboard fallback', () => {
		for (const pathname of routedAdminModuleSamples) {
			const crumbs = buildAdminBreadcrumbs(pathname, adminRoles, adminPermissions);
			expect.soft(crumbs[0]?.section, pathname).not.toBe('0');
			expect.soft(crumbs.every((crumb) => crumb.section), pathname).toBe(true);
		}
	});

	it('compacts long breadcrumb trails for narrow layouts without losing current context', () => {
		const crumbs = buildAdminBreadcrumbs('/tu/surat-keterangan/42/print', adminRoles, adminPermissions);
		expect(compactAdminBreadcrumbs(crumbs)).toEqual([
			{ label: 'Tata Usaha', href: undefined, section: '8' },
			{ label: 'Detail', href: undefined, section: '8.1.4.1' },
			{ label: 'Cetak', href: undefined, section: '8.1.4.1.1' }
		]);
	});
});
