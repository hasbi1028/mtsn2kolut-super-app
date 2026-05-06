export type SidebarNavItem = {
	href: string;
	label: string;
	icon: string;
	roles?: string[];
	permissions?: string[];
	pinnable?: boolean;
};

export type SidebarNavGroup = {
	group: string;
	items: SidebarNavItem[];
};

export const sidebarNavGroups: SidebarNavGroup[] = [
	{
		group: 'Utama',
		items: [{ href: '/', label: 'Dashboard', icon: 'grid', pinnable: false }]
	},
	{
		group: 'Akademik & Pembelajaran',
		items: [
			{ href: '/academic', label: 'Data Akademik', icon: 'book-open', roles: ['admin'], permissions: ['academic.read'] },
			{ href: '/jadwal', label: 'Jadwal', icon: 'calendar', roles: ['guru', 'siswa', 'ortu'] },
			{ href: '/grades', label: 'Nilai', icon: 'clipboard', roles: ['admin', 'guru'] },
			{ href: '/grades/rapor', label: 'Cetak Rapor', icon: 'printer', roles: ['admin', 'guru'] },
			{ href: '/journal', label: 'Jurnal Kelas', icon: 'journal', roles: ['admin', 'guru'] }
		]
	},
	{
		group: 'Siswa & Wali',
		items: [
			{ href: '/students', label: 'Siswa', icon: 'users', roles: ['admin', 'kesiswaan', 'guru'], permissions: ['students.read'] },
			{ href: '/parents', label: 'Orang Tua', icon: 'user-group', roles: ['admin'], permissions: ['parents.read'] },
			{ href: '/kesiswaan', label: 'Kesiswaan', icon: 'user-check', roles: ['admin', 'kesiswaan', 'guru'], permissions: ['students.read'] }
		]
	},
	{
		group: 'Bank Soal',
		items: [
			{ href: '/bank-soal', label: 'Dashboard Bank Soal', icon: 'grid', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },
			{ href: '/bank-soal/daftar', label: 'Daftar Soal', icon: 'book-open', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },
			{ href: '/bank-soal/tambah', label: 'Tambah Soal', icon: 'pen-tool', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },
			{ href: '/bank-soal/verifikasi', label: 'Review Soal', icon: 'clipboard', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },
			{ href: '/bank-soal/impor', label: 'Impor Soal', icon: 'file-text', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },
			{ href: '/bank-soal/analisis-butir', label: 'Analisis Butir', icon: 'activity', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },
			{ href: '/bank-soal/mapel-kd', label: 'Mapel & KD', icon: 'layers', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] },
			{ href: '/bank-soal/pengaturan', label: 'Pengaturan Bank Soal', icon: 'settings', roles: ['admin', 'guru'], permissions: ['bank_soal.read'] }
		]
	},
	{
		group: 'Asesmen',
		items: [
			{ href: '/asesmen', label: 'Dashboard Asesmen', icon: 'grid', roles: ['admin', 'guru', 'staf'], permissions: ['asesmen.read'] },
			{ href: '/asesmen/paket', label: 'Paket Soal', icon: 'book-open', roles: ['admin'], permissions: ['asesmen.package_manage'] },
			{ href: '/asesmen/kegiatan', label: 'Kegiatan', icon: 'calendar', roles: ['admin'], permissions: ['asesmen.event_manage'] },
			{ href: '/asesmen/persiapan', label: 'Persiapan', icon: 'file-text', roles: ['admin', 'guru'], permissions: ['asesmen.read'] },
			{ href: '/asesmen/pelaksanaan', label: 'Pelaksanaan', icon: 'activity', roles: ['admin', 'guru', 'staf'], permissions: ['asesmen.proctor'] },
			{ href: '/asesmen/hasil', label: 'Hasil', icon: 'clipboard', roles: ['admin', 'guru'], permissions: ['asesmen.result_read'] }
		]
	},
	{
		group: 'Administrasi & TU',
		items: [
			{ href: '/governance', label: 'Tata Kelola', icon: 'layers', roles: ['admin', 'staf'], permissions: ['governance.read'] },
			{ href: '/governance/actions', label: 'Tindak Lanjut', icon: 'clipboard', roles: ['admin', 'staf'], permissions: ['governance.manage'] },
			{ href: '/document-cycles', label: 'Monitoring Dokumen', icon: 'calendar', roles: ['admin', 'staf'], permissions: ['document_cycles.read'] },
			{ href: '/document-cycles/verifikasi', label: 'Verifikasi Dokumen', icon: 'user-check', roles: ['admin', 'staf'], permissions: ['document_cycles.manage'] },
			{ href: '/tu', label: 'Dashboard TU', icon: 'clipboard', roles: ['admin', 'staf'], permissions: ['letters.read'] },
			{ href: '/tu/surat-masuk', label: 'Surat Masuk', icon: 'inbox', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
			{ href: '/tu/surat-keluar', label: 'Surat Keluar', icon: 'mail', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
			{ href: '/tu/surat-keterangan', label: 'Surat Keterangan', icon: 'file-text', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
			{ href: '/tu/arsip', label: 'Arsip', icon: 'archive', roles: ['admin', 'staf'], permissions: ['letters.read'] },
			{ href: '/tu/disposisi', label: 'Disposisi', icon: 'mail-forward', roles: ['admin', 'staf'], permissions: ['letters.manage'] },
			{ href: '/tu/compliance-pack', label: 'Paket Kepatuhan', icon: 'printer', roles: ['admin', 'staf'], permissions: ['letters.read'] }
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
		group: 'Pegawai & PUSAKA',
		items: [
			{ href: '/employees', label: 'Master Pegawai', icon: 'user-check', roles: ['admin'] },
			{ href: '/pusaka', label: 'Kontrol & Monitor', icon: 'server', roles: ['admin'], permissions: ['pusaka.read'] },
			{ href: '/pusaka/employees', label: 'Pegawai PUSAKA', icon: 'user-check', roles: ['admin'], permissions: ['pusaka.manage'] },
			{ href: '/pusaka/kehadiran', label: 'Data Kehadiran', icon: 'clock', roles: ['admin'], permissions: ['pusaka.read'] },
			{ href: '/pusaka/summary', label: 'Ringkasan Kehadiran', icon: 'layers', roles: ['admin'], permissions: ['pusaka.read'] },
			{ href: '/pusaka/antrian', label: 'Antrian Job', icon: 'activity', roles: ['admin'], permissions: ['pusaka.manage'] }
		]
	},
		{
			group: 'Sistem',
			items: [
				{ href: '/notifications', label: 'Notifikasi', icon: 'activity' },
				{ href: '/settings/account', label: 'Akun Saya', icon: 'user-check' },
				{ href: '/settings', label: 'Pengaturan Sistem', icon: 'settings', roles: ['admin'] },
				{ href: '/settings/users', label: 'Manajemen User', icon: 'users', roles: ['admin'], permissions: ['users.read'] },
				{ href: '/settings/audit-logs', label: 'Audit Trail', icon: 'file-text', roles: ['admin'], permissions: ['audit.read'] },
				{ href: '/settings/school-profile', label: 'Profil Madrasah', icon: 'settings', roles: ['admin'], permissions: ['settings.school_profile'] }
			]
		}
	];

export const defaultPinnedByRole: Record<string, string[]> = {
	admin: ['/asesmen/persiapan', '/grades', '/settings'],
	guru: ['/asesmen/pelaksanaan', '/asesmen/persiapan', '/grades', '/jadwal'],
	staf: ['/document-cycles', '/inventory', '/library'],
	kesiswaan: ['/kesiswaan', '/students']
};
