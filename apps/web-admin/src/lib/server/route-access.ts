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
	'/parents',
	'/pusaka',
	'/website',
	'/cbt/events',
	'/cbt/packages',
	'/cbt/sessions',
	'/settings/users',
	'/settings/audit-logs',
	'/settings/school-profile',
	'/api/academic',
	'/api/employees',
	'/api/parents',
	'/api/users',
	'/api/pusaka',
	'/api/website',
	'/api/school-profile',
	'/api/cbt/events',
	'/api/cbt/packages',
	'/api/cbt/sessions',
	'/api/scheduler/tick'
] as const;

const GURU_SAFE_CBT_SUPPORT_READ_PATHS = new Set([
	'/api/cbt/events',
	'/api/cbt/soal-support/subjects'
]);

const GURU_SAFE_CBT_SUPPORT_READ_PREFIXES = [
	'/api/cbt/events'
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
const STUDENT_PAGE_PREFIXES = ['/students'] as const;
const STUDENT_API_PREFIXES = ['/api/students'] as const;

function matchesPathSegment(pathname: string, prefix: string) {
	const normalizedPrefix = prefix === '/' ? '/' : prefix.replace(/\/$/, '');
	return pathname === normalizedPrefix || pathname.startsWith(`${normalizedPrefix}/`);
}

export function isPublicPath(pathname: string) {
	if (PUBLIC_EXACT_PATHS.has(pathname)) return true;
	return PUBLIC_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
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
	return ADMIN_ONLY_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}

export function isGuruSafeCbtSupportReadPath(pathname: string, method: string) {
	if (!isReadMethod(method)) return false;
	if (GURU_SAFE_CBT_SUPPORT_READ_PATHS.has(pathname)) return true;
	return GURU_SAFE_CBT_SUPPORT_READ_PREFIXES.some((prefix) =>
		matchesPathSegment(pathname, prefix) && pathname.endsWith('/question-targets')
	);
}

export function isStaffOperationPath(pathname: string) {
	return STAFF_OPERATION_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}

export function isKesiswaanPath(pathname: string) {
	return KESISWAAN_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}

export function isStudentPagePath(pathname: string) {
	return STUDENT_PAGE_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}

export function isStudentApiPath(pathname: string) {
	return STUDENT_API_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}

export function isReadMethod(method: string) {
	return method === 'GET' || method === 'HEAD' || method === 'OPTIONS';
}
