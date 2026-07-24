export type SidebarNavItem = {
	kind?: 'item' | 'folder';
	id?: string;
	href: string;
	label: string;
	icon: string;
	section?: string;
	numberedLabel?: string;
	roles?: string[];
	permissions: string[];
	roleFallbacks?: string[];
	allowAuthenticatedFallback?: boolean;
	pinnable?: boolean;
	children?: SidebarNavItem[];
};

export type SidebarNavNode = SidebarNavItem;

export type SidebarFlatItem = SidebarNavItem & {
	group: string;
	groupSection?: string;
	ancestors: string[];
	ancestorSections?: string[];
	breadcrumb: string[];
	section?: string;
	numberedLabel?: string;
};

export type SidebarNavGroup = {
	group: string;
	section?: string;
	numberedLabel?: string;
	items: SidebarNavNode[];
	// Note: SidebarNavNode is now just SidebarNavItem (folders removed)
};

export const dashboardNavItem: SidebarNavItem = {
	href: '/',
	label: 'Dashboard',
	icon: 'home',
	permissions: ['dashboard.read'],
	allowAuthenticatedFallback: true,
	pinnable: false
};

export const sidebarNavGroups: SidebarNavGroup[] = [
	{
		group: 'Beranda',
		items: [
			dashboardNavItem
		]
	},
	{
		group: 'Kehadiran',
		items: [
			{ href: '/pusaka', label: 'Monitor Kehadiran', icon: 'clock', roles: ['admin'], permissions: ['pusaka.read'] },
			{ href: '/pusaka/employees', label: 'Pegawai PUSAKA', icon: 'users', roles: ['admin'], permissions: ['pusaka.manage'] },
			{ href: '/pusaka/kehadiran', label: 'Data Kehadiran', icon: 'calendar', roles: ['admin'], permissions: ['pusaka.read'] },
			{ href: '/pusaka/summary', label: 'Ringkasan Kehadiran', icon: 'bar-chart', roles: ['admin'], permissions: ['pusaka.read'] },
			{ href: '/pusaka/telegram-laporan', label: 'Laporan Telegram', icon: 'send', roles: ['admin'], permissions: ['pusaka.manage'] },
			{ href: '/pusaka/antrian', label: 'Antrian Sinkronisasi', icon: 'refresh-cw', roles: ['admin'], permissions: ['pusaka.manage'] }
		]
	},
	{
		group: 'Pegawai',
		items: [
			{ href: '/employees', label: 'Daftar Pegawai', icon: 'users', roles: ['admin'], permissions: ['employees.read', 'employees.manage'] }
		]
	},
	{
		group: 'Akademik',
		items: [
			{ href: '/academic/semesters', label: 'Semester', icon: 'calendar', roles: ['admin'], permissions: ['academic.read'] },
			{ href: '/academic/rombels', label: 'Rombel', icon: 'users', roles: ['admin'], permissions: ['academic.read'] },
			{ href: '/academic/curriculum', label: 'Kurikulum', icon: 'book', roles: ['admin'], permissions: ['academic.read'] },
			{ href: '/academic/subject-assignments', label: 'Assign Guru', icon: 'users', roles: ['admin'], permissions: ['academic.read'] },
			{ href: '/academic/timetable', label: 'Jadwal Pelajaran', icon: 'calendar', roles: ['admin'], permissions: ['academic.read'] }
		]
	},
	{
		group: 'Kesiswaan',
		items: [
			{ href: '/kesiswaan/murid', label: 'Data Murid', icon: 'users', roles: ['admin'], permissions: ['kesiswaan.read'] }
		]
	},
	{
		group: 'Pengaturan',
		items: [
			{ href: '/settings/account', label: 'Akun Saya', icon: 'user', permissions: ['settings.account'], allowAuthenticatedFallback: true },
			{ href: '/settings/school-profile', label: 'Profil Madrasah', icon: 'building', roles: ['admin'], permissions: ['settings.school_profile'] },
			{ href: '/settings/users', label: 'Pengguna', icon: 'users', roles: ['admin'], permissions: ['users.read'] },
			{ href: '/settings/rbac', label: 'Hak Akses', icon: 'shield', roles: ['admin'], permissions: ['roles.read'] },
			{ href: '/settings/branding', label: 'Logo & Branding', icon: 'image', roles: ['admin'], permissions: ['settings.school_profile'] },
			{ href: '/settings/backups', label: 'Backup Data', icon: 'database', roles: ['admin'], permissions: ['backup.read'] },
			{ href: '/settings/audit-logs', label: 'Audit Aktivitas', icon: 'file-text', roles: ['admin'], permissions: ['audit.read'] },
			{ href: '/settings/analytics', label: 'Statistik', icon: 'bar-chart', roles: ['admin'], permissions: ['analytics.read'] }
		]
	}
];

export const defaultPinnedByRole: Record<string, string[]> = {
	admin: ['/', '/pusaka', '/employees'],
	guru: ['/'],
	staf: ['/'],
	kesiswaan: ['/'],
	siswa: ['/'],
	ortu: ['/']
};
