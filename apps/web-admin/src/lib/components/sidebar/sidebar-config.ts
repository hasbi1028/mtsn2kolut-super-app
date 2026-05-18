export type SidebarNavItem = {
	href: string;
	label: string;
	icon: string;
	roles?: string[];
	permissions: string[];
	roleFallbacks?: string[];
	allowAuthenticatedFallback?: boolean;
	pinnable?: boolean;
};

export type SidebarNavGroup = {
	group: string;
	items: SidebarNavItem[];
};

export const dashboardNavItem: SidebarNavItem = {
	href: '/',
	label: 'Dashboard',
	icon: 'grid',
	permissions: ['dashboard.read'],
	allowAuthenticatedFallback: true,
	pinnable: false
};

export const sidebarNavGroups: SidebarNavGroup[] = [
	{
		group: 'Beranda',
		items: [
			{ href: '/notifications', label: 'Notifikasi', icon: 'activity', permissions: ['notifications.read'] },
			{ href: '/akademik/kesiapan', label: 'Kesiapan Akademik & Rapor', icon: 'clipboard', roles: ['admin', 'guru'], permissions: ['academic.read'] }
		]
	},
	{
		group: 'Portal',
		items: [
			{ href: '/portal/siswa', label: 'Portal Siswa', icon: 'book-open', roles: ['siswa'], roleFallbacks: ['siswa'], permissions: ['student_portal.read'] },
			{ href: '/portal/orang-tua', label: 'Portal Orang Tua', icon: 'user-group', roles: ['ortu'], roleFallbacks: ['ortu'], permissions: ['parent_portal.read'] }
		]
	},
	{
		group: 'Akademik',
		items: [
			{ href: '/akademik', label: 'Ringkasan Akademik', icon: 'grid', roles: ['admin', 'guru', 'kesiswaan'], permissions: ['academic.read'] },
			{ href: '/akademik/tahun-ajaran', label: 'Tahun Ajaran', icon: 'calendar', roles: ['admin'], permissions: ['academic.read'] },
			{ href: '/akademik/kurikulum', label: 'Struktur Kurikulum', icon: 'book-open', roles: ['admin', 'guru'], permissions: ['academic.read'] },
			{ href: '/akademik/rombel', label: 'Rombel', icon: 'layers', roles: ['admin', 'guru', 'kesiswaan'], permissions: ['academic.read'] },
			{ href: '/akademik/mapel', label: 'Mapel', icon: 'book-open', roles: ['admin', 'guru'], permissions: ['academic.read'] },
			{ href: '/akademik/guru-mapel', label: 'Guru Mapel', icon: 'user-check', roles: ['admin', 'guru'], permissions: ['academic.read'] },
			{ href: '/akademik/jam-pelajaran', label: 'Jam Pelajaran', icon: 'calendar', roles: ['admin', 'guru'], permissions: ['academic.read'] },
			{ href: '/akademik/jadwal', label: 'Jadwal', icon: 'calendar', roles: ['admin', 'guru'], permissions: ['academic.read'] },
			{ href: '/akademik/beban-guru', label: 'Beban Guru', icon: 'activity', roles: ['admin', 'guru'], permissions: ['academic.read'] },
			{ href: '/journal', label: 'Jurnal Kelas', icon: 'journal', roles: ['admin', 'guru'], permissions: ['journal.read', 'journal.manage', 'journal.read_all', 'journal.manage_all'] }
		]
	},
	{
		group: 'Siswa & Orang Tua',
		items: [
			{ href: '/students', label: 'Siswa', icon: 'users', roles: ['admin', 'kesiswaan', 'guru'], permissions: ['students.read'] },
			{ href: '/parents', label: 'Orang Tua', icon: 'user-group', roles: ['admin'], permissions: ['parents.read'] },
			{ href: '/kesiswaan', label: 'Kesiswaan', icon: 'user-check', roles: ['admin', 'kesiswaan', 'guru'], permissions: ['kesiswaan.read', 'students.read'] }
		]
	},
	{
		group: 'Nilai & Rapor',
		items: [
			{ href: '/grades', label: 'Input Nilai', icon: 'clipboard', roles: ['admin', 'guru'], permissions: ['grades.read', 'grades.manage'] },
			{ href: '/grades/rapor', label: 'Rapor Siswa', icon: 'printer', roles: ['admin', 'guru'], permissions: ['grades.read', 'grades.manage'] }
		]
	},
	{
		group: 'Bank Soal',
		items: [
			{ href: '/bank-soal', label: 'Dashboard Bank Soal', icon: 'grid', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },
			{ href: '/bank-soal/daftar', label: 'Kelola Soal', icon: 'book-open', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },

			{ href: '/bank-soal/verifikasi', label: 'Review & Terbitkan', icon: 'clipboard', roles: ['admin'], permissions: ['bank_soal.review', 'bank_soal.publish'] },
			{ href: '/bank-soal/laporan', label: 'Laporan', icon: 'activity', roles: ['admin', 'guru'], permissions: ['bank_soal.read', 'bank_soal.analytics', 'bank_soal.review'] },
			{ href: '/bank-soal/analisis-butir', label: 'Mutu Soal', icon: 'activity', roles: ['admin', 'guru'], permissions: ['bank_soal.analytics'] },
			{ href: '/bank-soal/pengaturan', label: 'Pengaturan', icon: 'settings', roles: ['admin'], permissions: ['bank_soal.settings'] }
		]
	},
	{
		group: 'Asesmen Ujian',
		items: [
			{ href: '/asesmen', label: 'Hari Ini / Dashboard Ujian', icon: 'grid', roles: ['admin', 'guru', 'staf'], permissions: ['asesmen.read', 'asesmen.proctor', 'asesmen.result_read', 'asesmen.score'] },
			{ href: '/asesmen/persiapan', label: 'Persiapan Ujian', icon: 'file-text', roles: ['admin', 'guru'], permissions: ['asesmen.read'] },
			{ href: '/asesmen/pelaksanaan', label: 'Pemantauan Ujian', icon: 'activity', roles: ['admin', 'guru', 'staf'], permissions: ['asesmen.proctor'] },
			{ href: '/asesmen/hasil', label: 'Hasil & BA', icon: 'clipboard', roles: ['admin', 'guru'], permissions: ['asesmen.result_read'] },
			{ href: '/asesmen/kegiatan', label: 'Arsip', icon: 'archive', roles: ['admin', 'guru'], permissions: ['asesmen.read', 'asesmen.event_manage'] },
			{ href: '/asesmen/aplikasi-siswa', label: 'Aplikasi Siswa', icon: 'package', roles: ['admin', 'guru', 'staf'], permissions: ['asesmen.read', 'asesmen.proctor'] }
		]
	},
	{
		group: 'Tata Usaha',
		items: [
			{ href: '/tu', label: 'Dashboard TU', icon: 'clipboard', roles: ['admin', 'staf'], permissions: ['letters.read'] },
			{ href: '/tu/surat-masuk', label: 'Surat Masuk', icon: 'inbox', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
			{ href: '/tu/surat-keluar', label: 'Surat Keluar', icon: 'mail', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
			{ href: '/tu/surat-keterangan', label: 'Surat Keterangan', icon: 'file-text', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
			{ href: '/tu/disposisi', label: 'Disposisi', icon: 'mail-forward', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
			{ href: '/tu/arsip', label: 'Arsip', icon: 'archive', roles: ['admin', 'staf'], permissions: ['archives.read'] },
			{ href: '/tu/compliance-pack', label: 'Paket Kepatuhan', icon: 'printer', roles: ['admin', 'staf'], permissions: ['letters.read'] },
			{ href: '/document-cycles', label: 'Monitoring Dokumen', icon: 'calendar', roles: ['admin', 'staf'], permissions: ['document_cycles.read'] },
			{ href: '/document-cycles/verifikasi', label: 'Verifikasi Dokumen', icon: 'user-check', roles: ['admin', 'staf'], permissions: ['document_cycles.manage'] },
			{ href: '/governance/actions', label: 'Tindak Lanjut', icon: 'clipboard', roles: ['admin', 'staf'], permissions: ['governance.manage'] },
			{ href: '/governance', label: 'Tata Kelola', icon: 'layers', roles: ['admin', 'staf'], permissions: ['governance.read'] }
		]
	},
	{
		group: 'Aset & Layanan',
		items: [
			{ href: '/library', label: 'Perpustakaan', icon: 'book-open', roles: ['admin', 'staf'], permissions: ['library.read'] },
			{ href: '/library/books', label: 'Katalog Buku', icon: 'book', roles: ['admin', 'staf'], permissions: ['library.manage'] },
			{ href: '/library/loans', label: 'Peminjaman', icon: 'repeat', roles: ['admin', 'staf'], permissions: ['library.manage'] },
			{ href: '/inventory', label: 'Inventaris', icon: 'package', roles: ['admin', 'staf'], permissions: ['inventory.read'] },
			{ href: '/inventory/items', label: 'Daftar Barang', icon: 'layers', roles: ['admin', 'staf'], permissions: ['inventory.manage'] }
		]
	},
	{
		group: 'Website',
		items: [
			{ href: '/website', label: 'Website Publik', icon: 'globe', roles: ['admin'], permissions: ['website.read'] },
			{ href: '/website/posts', label: 'Berita', icon: 'file-text', roles: ['admin'], permissions: ['website.manage'] },
			{ href: '/website/announcements', label: 'Pengumuman', icon: 'clipboard', roles: ['admin'], permissions: ['website.manage'] },
			{ href: '/website/pages', label: 'Halaman Publik', icon: 'book-open', roles: ['admin'], permissions: ['website.manage'] }
		]
	},
	{
		group: 'Pegawai & Kehadiran',
		items: [
			{ href: '/employees', label: 'Master Pegawai', icon: 'user-check', roles: ['admin'], permissions: ['employees.read', 'employees.manage'] },
			{ href: '/pusaka', label: 'Monitor PUSAKA', icon: 'server', roles: ['admin'], permissions: ['pusaka.read'] },
			{ href: '/pusaka/employees', label: 'Pegawai PUSAKA', icon: 'user-check', roles: ['admin'], permissions: ['pusaka.manage'] },
			{ href: '/pusaka/kehadiran', label: 'Data Kehadiran', icon: 'clock', roles: ['admin'], permissions: ['pusaka.read'] },
			{ href: '/pusaka/summary', label: 'Ringkasan Kehadiran', icon: 'layers', roles: ['admin'], permissions: ['pusaka.read'] },
			{ href: '/pusaka/telegram-laporan', label: 'Laporan Telegram', icon: 'send', roles: ['admin'], permissions: ['pusaka.manage'] },
			{ href: '/pusaka/antrian', label: 'Antrian Sinkronisasi', icon: 'activity', roles: ['admin'], permissions: ['pusaka.manage'] }
		]
	},
	{
		group: 'Pengaturan',
		items: [
			{ href: '/settings/account', label: 'Akun Saya', icon: 'user-check', permissions: ['settings.account'], allowAuthenticatedFallback: true },
			{ href: '/settings/school-profile', label: 'Profil Madrasah', icon: 'settings', roles: ['admin'], permissions: ['settings.school_profile'] },
			{ href: '/settings/branding', label: 'Logo & Branding', icon: 'settings', roles: ['admin'], permissions: ['settings.school_profile'] },
			{ href: '/settings/users', label: 'Pengguna & Hak Akses', icon: 'users', roles: ['admin'], permissions: ['users.read'] },
			{ href: '/settings/rbac', label: 'Peran & Izin Akses', icon: 'shield', roles: ['admin'], permissions: ['roles.read'] },
			{ href: '/settings/user-change-requests', label: 'Perubahan Data', icon: 'file-text', roles: ['admin'], permissions: ['profile_changes.review'] },
			{ href: '/settings/backups', label: 'Backup & Restore', icon: 'server', roles: ['admin'], permissions: ['backup.read'] },
			{ href: '/settings/maintenance', label: 'Maintenance Center', icon: 'shield', roles: ['admin'], permissions: ['roles.manage'] },
			{ href: '/settings/audit-logs', label: 'Audit Aktivitas', icon: 'file-text', roles: ['admin'], permissions: ['audit.read'] },
			{ href: '/settings/analytics', label: 'Statistik Penggunaan', icon: 'activity', roles: ['admin'], permissions: ['analytics.read'] },
			{ href: '/settings', label: 'Pengaturan Sistem', icon: 'settings', roles: ['admin'], permissions: ['settings.account'] }
		]
	}
];

export const defaultPinnedByRole: Record<string, string[]> = {
	admin: ['/akademik/kesiapan', '/akademik/rombel', '/akademik/jadwal', '/grades/rapor', '/asesmen/persiapan'],
	guru: ['/journal', '/grades', '/grades/rapor', '/bank-soal', '/akademik/jadwal'],
	staf: ['/document-cycles', '/inventory', '/library'],
	kesiswaan: ['/students', '/akademik/rombel', '/parents', '/kesiswaan'],
	siswa: ['/portal/siswa', '/jadwal'],
	ortu: ['/portal/orang-tua', '/jadwal']
};
