export type SidebarNavItem = {
	href: string;
	label: string;
	icon: string;
	roles?: string[];
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
		group: 'Akademik',
		items: [
			{ href: '/academic', label: 'Data Akademik', icon: 'book-open', roles: ['admin'] },
			{ href: '/jadwal', label: 'Jadwal', icon: 'calendar', roles: ['guru', 'siswa', 'ortu'] },
			{ href: '/grades', label: 'Nilai', icon: 'clipboard', roles: ['admin', 'guru'] },
			{ href: '/grades/rapor', label: 'Cetak Rapor', icon: 'printer', roles: ['admin', 'guru'] },
			{ href: '/journal', label: 'Jurnal Kelas', icon: 'journal', roles: ['admin', 'guru'] },
			{ href: '/students', label: 'Siswa', icon: 'users' },
			{ href: '/parents', label: 'Orang Tua', icon: 'user-group', roles: ['admin', 'staf'] }
		]
	},
	{
		group: 'CBT',
		items: [
			{ href: '/cbt/events', label: 'Kegiatan Ujian', icon: 'calendar', roles: ['admin'] },
			{ href: '/cbt/byod', label: 'Panduan BYOD', icon: 'activity', roles: ['admin', 'guru'] },
			{ href: '/cbt/questions', label: 'Bank Soal', icon: 'file-text' },
			{ href: '/cbt/soal', label: 'Komposer Soal', icon: 'pen-tool' },
			{ href: '/cbt/packages', label: 'Paket Ujian', icon: 'package' },
			{ href: '/cbt/sessions', label: 'Sesi Ujian', icon: 'play' }
		]
	},
	{
		group: 'Operasional',
		items: [{ href: '/employees', label: 'Master Pegawai', icon: 'user-check', roles: ['admin'] }]
	},
	{
		group: 'Perpustakaan',
		items: [
			{ href: '/library', label: 'Dashboard', icon: 'book-open', roles: ['admin', 'staf'] },
			{ href: '/library/books', label: 'Katalog Buku', icon: 'book', roles: ['admin', 'staf'] },
			{ href: '/library/loans', label: 'Peminjaman', icon: 'repeat', roles: ['admin', 'staf'] }
		]
	},
	{
		group: 'Inventaris',
		items: [
			{ href: '/inventory', label: 'Dashboard', icon: 'package', roles: ['admin', 'staf'] },
			{ href: '/inventory/items', label: 'Daftar Barang', icon: 'layers', roles: ['admin', 'staf'] }
		]
	},
	{
		group: 'Tata Usaha',
		items: [
			{ href: '/tu/surat-masuk', label: 'Surat Masuk', icon: 'inbox', roles: ['admin', 'staf'] },
			{ href: '/tu/surat-keluar', label: 'Surat Keluar', icon: 'mail', roles: ['admin', 'staf'] },
			{ href: '/tu/disposisi', label: 'Disposisi', icon: 'mail-forward', roles: ['admin', 'staf'] }
		]
	},
	{
		group: 'Website',
		items: [
			{ href: '/website', label: 'Website Publik', icon: 'globe', roles: ['admin'] },
			{ href: '/website/posts', label: 'Berita', icon: 'file-text', roles: ['admin'] },
			{ href: '/website/announcements', label: 'Pengumuman', icon: 'clipboard', roles: ['admin'] },
			{ href: '/website/pages', label: 'Halaman Publik', icon: 'book-open', roles: ['admin'] }
		]
	},
	{
		group: 'PUSAKA',
		items: [
			{ href: '/pusaka', label: 'Kontrol & Monitor', icon: 'server', roles: ['admin'] },
			{ href: '/pusaka/employees', label: 'Pegawai PUSAKA', icon: 'user-check', roles: ['admin'] },
			{ href: '/pusaka/kehadiran', label: 'Data Kehadiran', icon: 'clock', roles: ['admin'] },
			{ href: '/pusaka/summary', label: 'Ringkasan Kehadiran', icon: 'layers', roles: ['admin'] },
			{ href: '/pusaka/antrian', label: 'Antrian Job', icon: 'activity', roles: ['admin'] }
		]
	},
	{
		group: 'Sistem',
		items: [
			{ href: '/settings/users', label: 'Manajemen User', icon: 'users', roles: ['admin'] },
			{ href: '/settings/audit-logs', label: 'Audit Trail', icon: 'file-text', roles: ['admin'] },
			{ href: '/settings', label: 'Pengaturan', icon: 'settings', roles: ['admin'] }
		]
	}
];

export const defaultPinnedByRole: Record<string, string[]> = {
	admin: ['/cbt/sessions', '/grades', '/settings'],
	guru: ['/cbt/questions', '/grades', '/jadwal'],
	staf: ['/inventory', '/library']
};
