export type SidebarNavItem = {
	kind?: 'item';
	href: string;
	label: string;
	icon: string;
	roles?: string[];
	permissions: string[];
	roleFallbacks?: string[];
	allowAuthenticatedFallback?: boolean;
	pinnable?: boolean;
};

export type SidebarFolderItem = {
	kind: 'folder';
	id: string;
	label: string;
	icon: string;
	permissions?: string[];
	roleFallbacks?: string[];
	allowAuthenticatedFallback?: boolean;
	children: SidebarNavNode[];
	pinnable?: false;
};

export type SidebarNavNode = SidebarNavItem | SidebarFolderItem;

export type SidebarFlatItem = SidebarNavItem & {
	group: string;
	ancestors: string[];
	breadcrumb: string[];
};

export type SidebarNavGroup = {
	group: string;
	items: SidebarNavNode[];
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
			{ kind: 'folder', id: 'ringkasan', label: 'Ringkasan', icon: 'grid', children: [
				{ href: '/notifications', label: 'Notifikasi', icon: 'activity', permissions: ['notifications.read'] },
				{ href: '/akademik/kesiapan', label: 'Kesiapan Akademik & Rapor', icon: 'clipboard', roles: ['admin', 'guru'], permissions: ['academic.read'] }
			]}
		]
	},
	{
		group: 'Portal',
		items: [
			{ kind: 'folder', id: 'portal-pengguna', label: 'Portal Pengguna', icon: 'book-open', children: [
				{ href: '/portal/siswa', label: 'Portal Siswa', icon: 'book-open', roles: ['siswa'], roleFallbacks: ['siswa'], permissions: ['student_portal.read'] },
				{ href: '/portal/siswa/qr-login', label: 'QR Login Siswa', icon: 'activity', roles: ['siswa'], roleFallbacks: ['siswa'], permissions: ['student_portal.read'] },
				{ href: '/portal/orang-tua', label: 'Portal Orang Tua', icon: 'user-group', roles: ['ortu'], roleFallbacks: ['ortu'], permissions: ['parent_portal.read'] },
				{ href: '/jadwal', label: 'Jadwal Saya', icon: 'calendar', roles: ['siswa', 'ortu', 'guru'], roleFallbacks: ['siswa', 'ortu'], permissions: ['student_portal.read', 'parent_portal.read', 'academic.read'] }
			]}
		]
	},
	{
		group: 'Akademik',
		items: [
			{ kind: 'folder', id: 'tahun-kurikulum', label: 'Data Tahun & Kurikulum', icon: 'book-open', children: [
				{ href: '/akademik', label: 'Ringkasan Akademik', icon: 'grid', roles: ['admin', 'guru', 'kesiswaan'], permissions: ['academic.read'] },
				{ href: '/academic', label: 'Dashboard Akademik Lama', icon: 'grid', roles: ['admin', 'guru', 'kesiswaan'], permissions: ['academic.read'] },
				{ href: '/akademik/tahun-ajaran', label: 'Tahun Ajaran', icon: 'calendar', roles: ['admin'], permissions: ['academic.read'] },
				{ href: '/akademik/kurikulum', label: 'Struktur Kurikulum', icon: 'book-open', roles: ['admin', 'guru'], permissions: ['academic.read'] },
				{ href: '/akademik/mapel', label: 'Mapel', icon: 'book-open', roles: ['admin', 'guru'], permissions: ['academic.read'] }
			]},
			{ kind: 'folder', id: 'rombel-pengajaran', label: 'Rombel & Pengajaran', icon: 'layers', children: [
				{ href: '/akademik/rombel', label: 'Rombel', icon: 'layers', roles: ['admin', 'guru', 'kesiswaan'], permissions: ['academic.read'] },
				{ href: '/akademik/guru-mapel', label: 'Guru Mapel', icon: 'user-check', roles: ['admin', 'guru'], permissions: ['academic.read'] },
				{ href: '/akademik/beban-guru', label: 'Beban Guru', icon: 'activity', roles: ['admin', 'guru'], permissions: ['academic.read'] }
			]},
			{ kind: 'folder', id: 'jadwal-jurnal', label: 'Jadwal & Jurnal', icon: 'calendar', children: [
				{ href: '/akademik/jam-pelajaran', label: 'Jam Pelajaran', icon: 'calendar', roles: ['admin', 'guru'], permissions: ['academic.read'] },
				{ href: '/akademik/jadwal', label: 'Jadwal', icon: 'calendar', roles: ['admin', 'guru'], permissions: ['academic.read'] },
				{ href: '/journal', label: 'Jurnal Kelas', icon: 'journal', roles: ['admin', 'guru'], permissions: ['journal.read', 'journal.manage', 'journal.read_all', 'journal.manage_all'] }
			]}
		]
	},
	{
		group: 'Siswa & Orang Tua',
		items: [
			{ kind: 'folder', id: 'data-siswa', label: 'Data Siswa', icon: 'users', children: [
				{ href: '/students', label: 'Siswa', icon: 'users', roles: ['admin', 'kesiswaan', 'guru'], permissions: ['students.read'] },
				{ href: '/parents', label: 'Orang Tua', icon: 'user-group', roles: ['admin'], permissions: ['parents.read'] },
				{ href: '/kesiswaan', label: 'Kesiswaan', icon: 'user-check', roles: ['admin', 'kesiswaan', 'guru'], permissions: ['kesiswaan.read', 'students.read'] }
			]},
			{ kind: 'folder', id: 'kartu-siswa', label: 'Kartu Siswa', icon: 'printer', children: [
				{ href: '/kesiswaan/kartu-siswa', label: 'Kartu Siswa', icon: 'printer', roles: ['admin', 'kesiswaan'], permissions: ['students.read'] },
				{ href: '/kesiswaan/kartu-siswa/scan', label: 'Scanner Kartu', icon: 'activity', roles: ['admin', 'kesiswaan', 'guru', 'staf'], permissions: ['students.read'] }
			]}
		]
	},
	{
		group: 'Nilai & Rapor',
		items: [
			{ kind: 'folder', id: 'penilaian', label: 'Penilaian', icon: 'clipboard', children: [
				{ href: '/grades', label: 'Input Nilai', icon: 'clipboard', roles: ['admin', 'guru'], permissions: ['grades.read', 'grades.manage'] }
			]},
			{ kind: 'folder', id: 'rapor', label: 'Rapor', icon: 'printer', children: [
				{ href: '/grades/rapor', label: 'Rapor Siswa', icon: 'printer', roles: ['admin', 'guru'], permissions: ['grades.read', 'grades.manage'] }
			]}
		]
	},
	{
		group: 'Bank Soal',
		items: [
			{ kind: 'folder', id: 'kelola-soal', label: 'Kelola Soal', icon: 'book-open', children: [
				{ href: '/bank-soal', label: 'Dashboard Bank Soal', icon: 'grid', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },
				{ href: '/bank-soal/daftar', label: 'Daftar Soal', icon: 'book-open', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },
				{ href: '/bank-soal/tambah', label: 'Tambah Soal', icon: 'file-text', roles: ['admin', 'guru'], permissions: ['bank_soal.create'] },
				{ href: '/bank-soal/impor', label: 'Impor Soal', icon: 'file-text', roles: ['admin', 'guru'], permissions: ['bank_soal.import', 'bank_soal.create'] },
				{ href: '/bank-soal/cetak', label: 'Cetak Soal', icon: 'printer', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] }
			]},
			{ kind: 'folder', id: 'review-penerbitan', label: 'Review & Penerbitan', icon: 'clipboard', children: [
				{ href: '/bank-soal/verifikasi', label: 'Verifikasi Soal', icon: 'clipboard', roles: ['admin'], permissions: ['bank_soal.review', 'bank_soal.publish'] },
				{ href: '/bank-soal/penerbitan', label: 'Penerbitan Soal', icon: 'shield', roles: ['admin'], permissions: ['bank_soal.publish'] }
			]},
			{ kind: 'folder', id: 'referensi-analisis', label: 'Referensi & Analisis', icon: 'activity', children: [
				{ href: '/bank-soal/mapel-kd', label: 'Mapel & KD', icon: 'layers', roles: ['admin', 'guru'], permissions: ['bank_soal.create', 'bank_soal.read'] },
				{ href: '/bank-soal/analisis-butir', label: 'Analisis Butir', icon: 'activity', roles: ['admin', 'guru'], permissions: ['bank_soal.analytics'] },
				{ href: '/bank-soal/laporan', label: 'Laporan Bank Soal', icon: 'activity', roles: ['admin', 'guru'], permissions: ['bank_soal.read', 'bank_soal.analytics', 'bank_soal.review'] },
				{ href: '/bank-soal/pengaturan', label: 'Pengaturan Bank Soal', icon: 'settings', roles: ['admin'], permissions: ['bank_soal.settings'] }
			]}
		]
	},
	{
		group: 'Asesmen Ujian',
		items: [
			{ href: '/asesmen/ringkas', label: 'Ringkas', icon: 'grid', roles: ['admin', 'guru', 'staf'], permissions: ['asesmen.read'] },
			{ href: '/asesmen/persiapan', label: 'Persiapan', icon: 'file-text', roles: ['admin', 'guru'], permissions: ['asesmen.read'] },
			{ href: '/asesmen/pelaksanaan', label: 'Pelaksanaan', icon: 'activity', roles: ['admin', 'guru', 'staf'], permissions: ['asesmen.proctor'] },
			{ href: '/asesmen/hasil', label: 'Hasil', icon: 'clipboard', roles: ['admin', 'guru'], permissions: ['asesmen.result_read'] }
		]
	},
	{
		group: 'Tata Usaha',
		items: [
			{ kind: 'folder', id: 'persuratan', label: 'Persuratan', icon: 'mail', children: [
				{ href: '/tu', label: 'Dashboard TU', icon: 'clipboard', roles: ['admin', 'staf'], permissions: ['letters.read'] },
				{ href: '/tu/surat-masuk', label: 'Surat Masuk', icon: 'inbox', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
				{ href: '/tu/surat-keluar', label: 'Surat Keluar', icon: 'mail', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
				{ href: '/tu/surat-keterangan', label: 'Surat Keterangan', icon: 'file-text', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
				{ href: '/tu/disposisi', label: 'Disposisi', icon: 'mail-forward', roles: ['admin', 'staf'], permissions: ['letters.manage'] }
			]},
			{ kind: 'folder', id: 'arsip-kepatuhan', label: 'Arsip & Kepatuhan', icon: 'archive', children: [
				{ href: '/tu/arsip', label: 'Arsip', icon: 'archive', roles: ['admin', 'staf'], permissions: ['archives.read'] },
				{ href: '/tu/compliance-pack', label: 'Paket Kepatuhan', icon: 'printer', roles: ['admin', 'staf'], permissions: ['letters.read'] },
				{ href: '/document-cycles', label: 'Monitoring Dokumen', icon: 'calendar', roles: ['admin', 'staf'], permissions: ['document_cycles.read'] },
				{ href: '/document-cycles/verifikasi', label: 'Verifikasi Dokumen', icon: 'user-check', roles: ['admin', 'staf'], permissions: ['document_cycles.manage'] }
			]},
			{ kind: 'folder', id: 'tata-kelola', label: 'Tata Kelola', icon: 'layers', children: [
				{ href: '/governance', label: 'Dashboard Tata Kelola', icon: 'layers', roles: ['admin', 'staf'], permissions: ['governance.read'] },
				{ href: '/governance/actions', label: 'Tindak Lanjut', icon: 'clipboard', roles: ['admin', 'staf'], permissions: ['governance.manage'] },
				{ href: '/governance/actions/calendar', label: 'Kalender', icon: 'calendar', roles: ['admin', 'staf'], permissions: ['governance.manage'] },
				{ href: '/governance/actions/meeting-pack', label: 'Paket Rapat', icon: 'printer', roles: ['admin', 'staf'], permissions: ['governance.manage'] },
				{ href: '/governance/print-pack', label: 'Cetak Paket Tata Kelola', icon: 'printer', roles: ['admin', 'staf'], permissions: ['governance.manage'] },
				{ href: '/governance/actions/owner-briefing', label: 'Briefing Owner', icon: 'file-text', roles: ['admin', 'staf'], permissions: ['governance.manage'] },
				{ href: '/governance/actions/snp-briefing', label: 'Briefing SNP', icon: 'file-text', roles: ['admin', 'staf'], permissions: ['governance.manage'] },
				{ href: '/governance/actions/evidence-briefing', label: 'Briefing Bukti', icon: 'file-text', roles: ['admin', 'staf'], permissions: ['governance.manage'] }
			]}
		]
	},
	{
		group: 'Aset & Layanan',
		items: [
			{ kind: 'folder', id: 'perpustakaan', label: 'Perpustakaan', icon: 'book-open', children: [
				{ href: '/library', label: 'Dashboard Perpustakaan', icon: 'book-open', roles: ['admin', 'staf'], permissions: ['library.read'] },
				{ href: '/library/books', label: 'Katalog Buku', icon: 'book', roles: ['admin', 'staf'], permissions: ['library.manage'] },
				{ href: '/library/loans', label: 'Peminjaman', icon: 'repeat', roles: ['admin', 'staf'], permissions: ['library.manage'] }
			]},
			{ kind: 'folder', id: 'inventaris', label: 'Inventaris', icon: 'package', children: [
				{ href: '/inventory', label: 'Dashboard Inventaris', icon: 'package', roles: ['admin', 'staf'], permissions: ['inventory.read'] },
				{ href: '/inventory/items', label: 'Daftar Barang', icon: 'layers', roles: ['admin', 'staf'], permissions: ['inventory.manage'] }
			]}
		]
	},
	{
		group: 'Website',
		items: [
			{ kind: 'folder', id: 'kelola-konten', label: 'Kelola Konten', icon: 'globe', children: [
				{ href: '/website', label: 'Dashboard Website', icon: 'globe', roles: ['admin'], permissions: ['website.read'] },
				{ href: '/berita', label: 'Berita Publik', icon: 'file-text', roles: ['admin'], permissions: ['website.read'] },
				{ href: '/pengumuman', label: 'Pengumuman Publik', icon: 'clipboard', roles: ['admin'], permissions: ['website.read'] },
				{ href: '/profil', label: 'Profil Publik', icon: 'book-open', roles: ['admin'], permissions: ['website.read'] },
				{ href: '/ppdb', label: 'PPDB Publik', icon: 'users', roles: ['admin'], permissions: ['website.read'] },
				{ href: '/kontak', label: 'Kontak Publik', icon: 'mail', roles: ['admin'], permissions: ['website.read'] },
				{ href: '/website/posts', label: 'Kelola Berita', icon: 'file-text', roles: ['admin'], permissions: ['website.manage'] },
				{ href: '/website/announcements', label: 'Pengumuman', icon: 'clipboard', roles: ['admin'], permissions: ['website.manage'] },
				{ href: '/website/pages', label: 'Halaman Publik', icon: 'book-open', roles: ['admin'], permissions: ['website.manage'] }
			]}
		]
	},
	{
		group: 'Pegawai & Kehadiran',
		items: [
			{ kind: 'folder', id: 'master-pegawai', label: 'Master Pegawai', icon: 'user-check', children: [
				{ href: '/employees', label: 'Master Pegawai', icon: 'user-check', roles: ['admin'], permissions: ['employees.read', 'employees.manage'] }
			]},
			{ kind: 'folder', id: 'pusaka', label: 'PUSAKA', icon: 'server', children: [
				{ href: '/pusaka', label: 'Monitor PUSAKA', icon: 'server', roles: ['admin'], permissions: ['pusaka.read'] },
				{ href: '/pusaka/employees', label: 'Pegawai PUSAKA', icon: 'user-check', roles: ['admin'], permissions: ['pusaka.manage'] },
				{ href: '/pusaka/kehadiran', label: 'Data Kehadiran', icon: 'clock', roles: ['admin'], permissions: ['pusaka.read'] },
				{ href: '/pusaka/summary', label: 'Ringkasan Kehadiran', icon: 'layers', roles: ['admin'], permissions: ['pusaka.read'] },
				{ href: '/pusaka/telegram-laporan', label: 'Laporan Telegram', icon: 'send', roles: ['admin'], permissions: ['pusaka.manage'] },
				{ href: '/pusaka/antrian', label: 'Antrian Sinkronisasi', icon: 'activity', roles: ['admin'], permissions: ['pusaka.manage'] }
			]}
		]
	},
	{
		group: 'Pengaturan',
		items: [
			{ kind: 'folder', id: 'akun-madrasah', label: 'Akun & Profil Madrasah', icon: 'settings', children: [
				{ href: '/settings/account', label: 'Akun Saya', icon: 'user-check', permissions: ['settings.account'], allowAuthenticatedFallback: true },
				{ href: '/settings/school-profile', label: 'Profil Madrasah', icon: 'settings', roles: ['admin'], permissions: ['settings.school_profile'] },
				{ href: '/settings/branding', label: 'Logo & Branding', icon: 'settings', roles: ['admin'], permissions: ['settings.school_profile'] }
			]},
			{ kind: 'folder', id: 'akses-keamanan', label: 'Pengguna & Hak Akses', icon: 'shield', children: [
				{ href: '/settings/users', label: 'Pengguna', icon: 'users', roles: ['admin'], permissions: ['users.read'] },
				{ href: '/settings/rbac', label: 'Peran & Izin Akses', icon: 'shield', roles: ['admin'], permissions: ['roles.read'] },
				{ href: '/settings/user-change-requests', label: 'Perubahan Data', icon: 'file-text', roles: ['admin'], permissions: ['profile_changes.review'] }
			]},
			{ kind: 'folder', id: 'sistem-audit', label: 'Sistem & Audit', icon: 'server', children: [
				{ href: '/settings', label: 'Pengaturan Sistem', icon: 'settings', roles: ['admin'], permissions: ['settings.account'] },
				{ href: '/maintenance', label: 'Mode Maintenance', icon: 'shield', roles: ['admin'], permissions: ['roles.manage'] },
				{ href: '/settings/backups', label: 'Backup & Restore', icon: 'server', roles: ['admin'], permissions: ['backup.read'] },
				{ href: '/settings/maintenance', label: 'Maintenance Center', icon: 'shield', roles: ['admin'], permissions: ['roles.manage'] },
				{ href: '/settings/audit-logs', label: 'Audit Aktivitas', icon: 'file-text', roles: ['admin'], permissions: ['audit.read'] },
				{ href: '/settings/analytics', label: 'Statistik Penggunaan', icon: 'activity', roles: ['admin'], permissions: ['analytics.read'] }
			]}
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
