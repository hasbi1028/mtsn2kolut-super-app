import type { AuthUser } from '$lib/server/auth';

export type MaintenanceMode = 'off' | 'global' | 'module' | 'read_only';
export type MaintenanceSeverity = 'info' | 'warning' | 'critical';

export type MaintenanceWindow = {
	id: string;
	title: string;
	message: string;
	mode: MaintenanceMode;
	affected_modules: string[];
	starts_at?: string;
	ends_at?: string;
	is_active: boolean;
	allow_admin_bypass: boolean;
	bypass_roles: string[];
	severity: MaintenanceSeverity | string;
	status: 'inactive' | 'scheduled' | 'active_now' | 'ended' | string;
	created_by?: string;
	updated_by?: string;
	created_at: string;
	updated_at: string;
};

export type MaintenanceStatus = {
	active: boolean;
	mode: MaintenanceMode;
	server_time: string;
	window?: MaintenanceWindow;
};

export const maintenanceModules = [
	{ code: 'global', label: 'Global' },
	{ code: 'auth', label: 'Login & sesi' },
	{ code: 'dashboard', label: 'Dashboard' },
	{ code: 'akademik', label: 'Akademik' },
	{ code: 'students', label: 'Siswa & orang tua' },
	{ code: 'cbt', label: 'Asesmen / CBT' },
	{ code: 'pusaka', label: 'PUSAKA' },
	{ code: 'backup_restore', label: 'Backup & restore' },
	{ code: 'settings', label: 'Pengaturan' }
] as const;

const routePrefixesByModule: Record<string, string[]> = {
	global: ['/'],
	auth: ['/login', '/settings/account', '/api/auth'],
	dashboard: ['/', '/notifications', '/api/notifications', '/api/internal-analytics'],
	akademik: ['/akademik', '/academic', '/journal', '/grades', '/jadwal', '/api/academic', '/api/journal', '/api/grades'],
	students: ['/students', '/parents', '/kesiswaan', '/portal', '/api/students', '/api/parents', '/api/kesiswaan', '/api/portal'],
	pusaka: ['/pusaka', '/api/pusaka'],
	backup_restore: ['/settings/backups', '/api/system/backups'],
	settings: ['/settings', '/api/settings', '/api/users', '/api/rbac', '/api/school-profile', '/api/branding', '/api/system/maintenance']
};

export function maintenanceModuleLabel(code: string) {
	return maintenanceModules.find((module) => module.code === code)?.label ?? code;
}

export function routeAffectedByMaintenance(window: MaintenanceWindow | undefined, pathname: string): boolean {
	if (!window) return false;
	if (window.mode === 'global' || window.affected_modules.includes('global')) return true;
	if (window.mode === 'read_only') return true;
	return window.affected_modules.some((module) =>
		(routePrefixesByModule[module] ?? []).some((prefix) => matchesPathSegment(pathname, prefix))
	);
}

export function canBypassMaintenance(user: AuthUser | null | undefined, window: MaintenanceWindow | undefined) {
	if (!user || !window?.allow_admin_bypass) return false;
	const roles = user.roles.length > 0 ? user.roles : user.role ? [user.role] : [];
	return window.bypass_roles.some((role) => roles.includes(role));
}

function matchesPathSegment(pathname: string, prefix: string) {
	if (prefix === '/') return pathname === '/';
	const normalized = prefix.replace(/\/$/, '');
	return pathname === normalized || pathname.startsWith(`${normalized}/`);
}
