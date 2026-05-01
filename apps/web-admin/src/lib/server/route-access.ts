import type { AuthUser } from '$lib/server/auth';

const PUBLIC_EXACT_PATHS = new Set([
	'/',
	'/login',
	'/ppdb',
	'/profil',
	'/berita',
	'/pengumuman',
	'/kontak',
	'/api/auth/logout',
	'/api/public/register-student',
	'/api/public/site/posts',
	'/api/public/site/announcements'
]);

const PUBLIC_PREFIXES = [
	'/berita/',
	'/pengumuman/',
	'/api/public/site/pages/',
	'/api/public/site/posts/',
	'/api/public/site/announcements/'
] as const;

const ADMIN_ONLY_PREFIXES = [
	'/employees',
	'/academic',
	'/pusaka',
	'/website',
	'/settings/users',
	'/settings/audit-logs',
	'/settings/school-profile',
	'/api/employees',
	'/api/users',
	'/api/pusaka',
	'/api/website'
] as const;

const STAFF_OPERATION_PREFIXES = [
	'/document-cycles',
	'/api/document-cycles',
	'/governance',
	'/api/governance',
	'/tu',
	'/api/tu',
	'/library',
	'/api/library',
	'/inventory',
	'/api/inventory'
] as const;

const KESISWAAN_PREFIXES = ['/kesiswaan', '/api/kesiswaan'] as const;

export function isPublicPath(pathname: string) {
	if (PUBLIC_EXACT_PATHS.has(pathname)) return true;
	return PUBLIC_PREFIXES.some((prefix) => pathname.startsWith(prefix));
}

export function userRoles(user: AuthUser | undefined): string[] {
	if (!user) return [];
	return user.roles.length > 0 ? user.roles : user.role ? [user.role] : [];
}

export function hasAnyRole(user: AuthUser | undefined, allowed: readonly string[]) {
	const roles = userRoles(user);
	return allowed.some((role) => roles.includes(role));
}

export function isAdminOnlyPath(pathname: string) {
	return ADMIN_ONLY_PREFIXES.some((prefix) => pathname.startsWith(prefix));
}

export function isStaffOperationPath(pathname: string) {
	return STAFF_OPERATION_PREFIXES.some((prefix) => pathname.startsWith(prefix));
}

export function isKesiswaanPath(pathname: string) {
	return KESISWAAN_PREFIXES.some((prefix) => pathname.startsWith(prefix));
}
