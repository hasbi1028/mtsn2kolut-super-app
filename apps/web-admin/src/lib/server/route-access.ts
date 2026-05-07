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
	'/asesmen/kegiatan',
	'/asesmen/paket',
	'/asesmen/sesi',
	'/settings/users',
	'/settings/rbac',
	'/settings/user-change-requests',
	'/settings/audit-logs',
	'/settings/school-profile',
	'/api/academic',
	'/api/employees',
	'/api/parents',
	'/api/users',
	'/api/pusaka',
	'/api/rbac',
	'/api/website',
	'/api/school-profile',
	'/api/asesmen/events',
	'/api/asesmen/packages',
	'/api/asesmen/sessions',
	'/api/scheduler/tick'
] as const;

const GURU_SAFE_ASSESSMENT_SUPPORT_READ_PATHS = new Set([
	'/asesmen/kegiatan',
	'/asesmen/paket',
	'/asesmen/sesi',
	'/api/asesmen/events',
	'/api/asesmen/packages',
	'/api/asesmen/sessions',
	'/api/bank-soal/soal-support/subjects'
]);

const GURU_SAFE_ASSESSMENT_SUPPORT_READ_PREFIXES = [
	'/asesmen/kegiatan',
	'/asesmen/paket',
	'/asesmen/sesi',
	'/api/asesmen/events',
	'/api/asesmen/packages',
	'/api/asesmen/sessions'
] as const;

const BANK_SOAL_PREFIXES = ['/bank-soal', '/api/bank-soal'] as const;

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
const STUDENT_PORTAL_PREFIXES = ['/portal/siswa', '/api/portal/siswa', '/api/portal/student'] as const;
const PARENT_PORTAL_PREFIXES = ['/portal/orang-tua', '/api/portal/orang-tua', '/api/portal/parent'] as const;

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

export function userPermissions(user: AuthUser | undefined): string[] {
	if (!user) return [];
	return (user.permissions ?? []).map((permission) => permission.trim()).filter(Boolean);
}

export function hasAnyRole(user: AuthUser | undefined, allowed: readonly string[]) {
	const roles = userRoles(user);
	return allowed.some((role) => roles.includes(role));
}

export function hasAnyPermission(user: AuthUser | undefined, allowed: readonly string[]) {
	const permissions = userPermissions(user);
	return allowed.some((permission) => permissions.includes(permission));
}

export function isAdminOnlyPath(pathname: string) {
	return ADMIN_ONLY_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}

export function isGuruSafeAssessmentSupportReadPath(pathname: string, method: string) {
	if (!isReadMethod(method)) return false;
	if (GURU_SAFE_ASSESSMENT_SUPPORT_READ_PATHS.has(pathname)) return true;
	return GURU_SAFE_ASSESSMENT_SUPPORT_READ_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}

export function isBankSoalPath(pathname: string) {
	return BANK_SOAL_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
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

export function isStudentPortalPath(pathname: string) {
	return STUDENT_PORTAL_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}

export function isParentPortalPath(pathname: string) {
	return PARENT_PORTAL_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}

export function isReadMethod(method: string) {
	return method === 'GET' || method === 'HEAD' || method === 'OPTIONS';
}

export function isMustChangePasswordAllowedPath(pathname: string, method: string) {
	if (isPublicPath(pathname)) return true;
	if (pathname.startsWith('/_app/') || pathname === '/favicon.svg' || pathname === '/manifest.webmanifest') return true;
	if (pathname === '/settings/account' && isReadMethod(method)) return true;
	if (pathname === '/api/auth/account' && isReadMethod(method)) return true;
	if (pathname === '/api/auth/change-password' && method === 'POST') return true;
	if (pathname === '/api/auth/logout' && method === 'POST') return true;
	if (pathname === '/api/auth/logout-all' && method === 'POST') return true;
	if (pathname === '/api/auth/sessions' && isReadMethod(method)) return true;
	if (matchesPathSegment(pathname, '/api/auth/sessions') && (isReadMethod(method) || method === 'DELETE')) return true;
	return false;
}

function usersPermission(pathname: string, method: string): string[] | undefined {
	if (matchesPathSegment(pathname, '/api/users/student-accounts')) return ['student_accounts.manage'];
	if (matchesPathSegment(pathname, '/api/users/parent-accounts')) return ['parent_accounts.manage'];
	if (matchesPathSegment(pathname, '/api/users/change-requests')) return ['profile_changes.review'];
	if (matchesPathSegment(pathname, '/api/users') && pathname.endsWith('/reset-password')) return ['users.reset_password'];
	if (matchesPathSegment(pathname, '/api/users') && pathname.endsWith('/profile-link')) return ['users.update'];
	if (matchesPathSegment(pathname, '/api/users') && pathname.endsWith('/roles')) return ['users.manage_roles'];
	if (matchesPathSegment(pathname, '/settings/user-change-requests')) return ['profile_changes.review'];
	if (matchesPathSegment(pathname, '/settings/users')) return ['users.read'];
	if (matchesPathSegment(pathname, '/api/users')) {
		if (isReadMethod(method)) return ['users.read'];
		if (method === 'POST') return ['users.create'];
		if (method === 'DELETE') return ['users.deactivate'];
		if (method === 'PATCH' || method === 'PUT') return ['users.update', 'users.deactivate', 'users.manage_roles'];
	}
	return undefined;
}

function bankSoalPermission(pathname: string, method: string): string[] | undefined {
	if (!isBankSoalPath(pathname)) return undefined;
	if (!pathname.startsWith('/api/')) {
		if (matchesPathSegment(pathname, '/bank-soal/tambah')) return ['bank_soal.create'];
		if (matchesPathSegment(pathname, '/bank-soal/verifikasi')) return ['bank_soal.review'];
		if (matchesPathSegment(pathname, '/bank-soal/impor')) return ['bank_soal.import'];
		if (matchesPathSegment(pathname, '/bank-soal/pengaturan')) return ['bank_soal.settings'];
		return ['bank_soal.read'];
	}
	if (isReadMethod(method)) return ['bank_soal.read'];
	if (matchesPathSegment(pathname, '/api/bank-soal/questions/import-legacy')) return ['bank_soal.import'];
	if (matchesPathSegment(pathname, '/api/bank-soal/questions/bulk-workflow')) return ['bank_soal.update', 'bank_soal.review', 'bank_soal.publish'];
	if (matchesPathSegment(pathname, '/api/bank-soal/assets') && method === 'POST') return ['bank_soal.create', 'bank_soal.update'];
	if (matchesPathSegment(pathname, '/api/bank-soal/questions') && pathname.endsWith('/duplicate') && method === 'POST') return ['bank_soal.create'];
	if (matchesPathSegment(pathname, '/api/bank-soal/questions') && pathname.endsWith('/revision') && method === 'POST') return ['bank_soal.update'];
	if (matchesPathSegment(pathname, '/api/bank-soal/questions') && pathname.endsWith('/workflow')) return ['bank_soal.update', 'bank_soal.review', 'bank_soal.publish'];
	if (method === 'POST') return ['bank_soal.create'];
	if (method === 'DELETE') return ['bank_soal.delete'];
	if (method === 'PATCH' || method === 'PUT') return ['bank_soal.update', 'bank_soal.review', 'bank_soal.publish'];
	return ['bank_soal.read'];
}

function isBankSoalGuruFallbackPath(pathname: string, method: string) {
	if (!isReadMethod(method)) return false;
	if (pathname.startsWith('/api/')) return true;
	return !matchesPathSegment(pathname, '/bank-soal/tambah')
		&& !matchesPathSegment(pathname, '/bank-soal/verifikasi')
		&& !matchesPathSegment(pathname, '/bank-soal/impor')
		&& !matchesPathSegment(pathname, '/bank-soal/pengaturan');
}

function asesmenPermission(pathname: string, method: string): string[] | undefined {
	if (matchesPathSegment(pathname, '/asesmen/hasil')) return ['asesmen.result_read'];
	if (matchesPathSegment(pathname, '/asesmen/pelaksanaan') || matchesPathSegment(pathname, '/asesmen/pengawasan')) return ['asesmen.proctor'];
	if (matchesPathSegment(pathname, '/asesmen/paket') || matchesPathSegment(pathname, '/api/asesmen/packages')) {
		return isReadMethod(method) && pathname.startsWith('/api/') ? ['asesmen.read'] : ['asesmen.package_manage'];
	}
	if (matchesPathSegment(pathname, '/asesmen/kegiatan') || matchesPathSegment(pathname, '/api/asesmen/events')) {
		return isReadMethod(method) && pathname.startsWith('/api/') ? ['asesmen.read'] : ['asesmen.event_manage'];
	}
	if (matchesPathSegment(pathname, '/asesmen/sesi') || matchesPathSegment(pathname, '/api/asesmen/sessions')) {
		return isReadMethod(method) && pathname.startsWith('/api/') ? ['asesmen.read'] : ['asesmen.proctor'];
	}
	if (matchesPathSegment(pathname, '/asesmen')) return ['asesmen.read'];
	return undefined;
}

function staffOperationPermission(pathname: string, method: string): string[] | undefined {
	if (!isStaffOperationPath(pathname)) return undefined;
	if (matchesPathSegment(pathname, '/library') || matchesPathSegment(pathname, '/api/library')) return isReadMethod(method) ? ['library.read'] : ['library.manage'];
	if (matchesPathSegment(pathname, '/inventory') || matchesPathSegment(pathname, '/api/inventory')) return isReadMethod(method) ? ['inventory.read'] : ['inventory.manage'];
	if (matchesPathSegment(pathname, '/document-cycles') || matchesPathSegment(pathname, '/api/document-cycles')) return isReadMethod(method) ? ['document_cycles.read'] : ['document_cycles.manage'];
	if (matchesPathSegment(pathname, '/governance') || matchesPathSegment(pathname, '/api/governance')) return isReadMethod(method) ? ['governance.read'] : ['governance.manage'];
	if (matchesPathSegment(pathname, '/tu') || matchesPathSegment(pathname, '/api/tu')) return isReadMethod(method) ? ['letters.read'] : ['letters.manage'];
	return undefined;
}

function isRombelTimetableJournalSessionPath(pathname: string): boolean {
	return /^\/api\/academic\/rombel\/[^/]+\/timetable-slots\/[^/]+\/journal-session\/?$/.test(pathname.split('?')[0] ?? pathname);
}

export function requiredPermissionsForPath(pathname: string, method: string): string[] {
	if (matchesPathSegment(pathname, '/settings/audit-logs')) return ['audit.read'];
	if (matchesPathSegment(pathname, '/settings/rbac')) return ['roles.read'];
	if (matchesPathSegment(pathname, '/settings/school-profile') || matchesPathSegment(pathname, '/api/school-profile')) return ['settings.school_profile'];
	if (matchesPathSegment(pathname, '/api/rbac')) return isReadMethod(method) ? ['roles.read'] : ['roles.manage'];
	if (matchesPathSegment(pathname, '/parents') || matchesPathSegment(pathname, '/api/parents')) return isReadMethod(method) ? ['parents.read'] : ['parents.manage'];
	if (isRombelTimetableJournalSessionPath(pathname)) return ['journal.manage', 'journal.manage_all'];
	if (matchesPathSegment(pathname, '/academic') || matchesPathSegment(pathname, '/api/academic')) return isReadMethod(method) ? ['academic.read'] : ['academic.manage'];
	if (matchesPathSegment(pathname, '/pusaka') || matchesPathSegment(pathname, '/api/pusaka')) return isReadMethod(method) ? ['pusaka.read'] : ['pusaka.manage'];
	if (matchesPathSegment(pathname, '/website') || matchesPathSegment(pathname, '/api/website')) return isReadMethod(method) ? ['website.read'] : ['website.manage'];
	if (isStudentPortalPath(pathname)) return ['student_portal.read'];
	if (isParentPortalPath(pathname)) return ['parent_portal.read'];
	if (isKesiswaanPath(pathname)) return isReadMethod(method) ? ['students.read'] : ['students.manage'];
	if (isStudentPagePath(pathname) || isStudentApiPath(pathname)) return isReadMethod(method) ? ['students.read'] : ['students.manage'];
	return usersPermission(pathname, method)
		?? bankSoalPermission(pathname, method)
		?? asesmenPermission(pathname, method)
		?? staffOperationPermission(pathname, method)
		?? [];
}

export function canAccessProtectedRoute(user: AuthUser | undefined, pathname: string, method: string): boolean {
	if (isPublicPath(pathname)) return true;
	if (!user) return false;
	if (user.must_change_password && !isMustChangePasswordAllowedPath(pathname, method)) return false;
	if (hasAnyRole(user, ['admin'])) return true;

	const requiredPermissions = requiredPermissionsForPath(pathname, method);
	if (requiredPermissions.length > 0 && hasAnyPermission(user, requiredPermissions)) return true;
	if (isRombelTimetableJournalSessionPath(pathname)) return hasAnyRole(user, ['guru']);

	if (isAdminOnlyPath(pathname) && !isGuruSafeAssessmentSupportReadPath(pathname, method)) return false;
	if (isBankSoalPath(pathname)) return isBankSoalGuruFallbackPath(pathname, method) && hasAnyRole(user, ['guru']);
	if (isStaffOperationPath(pathname)) return hasAnyRole(user, ['staf']);
	if (isStudentPortalPath(pathname)) return hasAnyRole(user, ['siswa']);
	if (isParentPortalPath(pathname)) return hasAnyRole(user, ['ortu']);
	if (isKesiswaanPath(pathname)) return hasAnyRole(user, isReadMethod(method) ? ['kesiswaan', 'guru'] : ['kesiswaan']);
	if (isStudentPagePath(pathname)) return hasAnyRole(user, ['kesiswaan']);
	if (isStudentApiPath(pathname)) return hasAnyRole(user, isReadMethod(method) ? ['kesiswaan', 'guru'] : ['kesiswaan']);
	return true;
}
