import type { AuthUser } from '$lib/server/auth';
import { isPublicPath, matchesPathSegment } from '$lib/routes/public-policy';

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
	'/settings/branding',
	'/settings/backups',
	'/settings/maintenance',
	'/api/academic',
	'/api/employees',
	'/api/parents',
	'/api/users',
	'/api/pusaka',
	'/api/rbac',
	'/api/website',
	'/api/school-profile',
	'/api/branding',
	'/api/system/backups',
	'/api/system/maintenance',
	'/api/asesmen/events',
	'/api/asesmen/approvals',
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
	'/api/asesmen/packages'
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
const SCHEDULE_PAGE_PREFIXES = ['/jadwal'] as const;
const GRADES_PREFIXES = ['/grades', '/api/grades'] as const;
const JOURNAL_PREFIXES = ['/journal', '/api/journal'] as const;
const EMPLOYEE_PREFIXES = ['/employees', '/api/employees'] as const;

export { isPublicPath };

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
	if (isSensitiveAssessmentReadPath(pathname)) return false;
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
	if (matchesPathSegment(pathname, '/api/users/profile-candidates')) return ['users.create'];
	if (matchesPathSegment(pathname, '/api/users/change-requests')) return ['profile_changes.review'];
	if (matchesPathSegment(pathname, '/api/users') && pathname.endsWith('/reset-password')) return ['users.reset_password'];
	if (matchesPathSegment(pathname, '/api/users') && pathname.endsWith('/force-password-change')) return ['users.reset_password'];
	if (matchesPathSegment(pathname, '/api/users') && pathname.endsWith('/status')) return ['users.deactivate'];
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

function settingsPermission(pathname: string): string[] | undefined {
	if (matchesPathSegment(pathname, '/settings/account')) return undefined;
	if (matchesPathSegment(pathname, '/settings/audit-logs')) return ['audit.read'];
	if (matchesPathSegment(pathname, '/settings/analytics')) return ['analytics.read'];
	if (matchesPathSegment(pathname, '/settings/backups')) return ['backup.read'];
	if (matchesPathSegment(pathname, '/settings/maintenance') || matchesPathSegment(pathname, '/api/system/maintenance')) return ['settings.maintenance'];
	if (matchesPathSegment(pathname, '/settings/rbac')) return ['roles.read'];
	if (matchesPathSegment(pathname, '/settings/school-profile') || matchesPathSegment(pathname, '/api/school-profile')) return ['settings.school_profile'];
	if (matchesPathSegment(pathname, '/settings/branding') || matchesPathSegment(pathname, '/api/branding')) return ['settings.branding'];
	if (pathname === '/settings') return ['settings.account'];
	return undefined;
}

function bankSoalPermission(pathname: string, method: string): string[] | undefined {
	if (!isBankSoalPath(pathname)) return undefined;
	if (!pathname.startsWith('/api/')) {
		if (matchesPathSegment(pathname, '/bank-soal/tambah')) return ['bank_soal.create'];
		if (matchesPathSegment(pathname, '/bank-soal/verifikasi')) return ['bank_soal.review'];
		if (matchesPathSegment(pathname, '/bank-soal/penerbitan')) return ['bank_soal.publish'];
		if (matchesPathSegment(pathname, '/bank-soal/impor')) return ['bank_soal.import'];
		if (matchesPathSegment(pathname, '/bank-soal/pengaturan')) return ['bank_soal.settings'];
		if (matchesPathSegment(pathname, '/bank-soal/analisis-butir')) return ['bank_soal.analytics'];
		return ['bank_soal.read'];
	}
	if (matchesPathSegment(pathname, '/api/bank-soal/reviewer-scopes')) return ['bank_soal.assign_reviewer', 'bank_soal.settings'];
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
		&& !matchesPathSegment(pathname, '/bank-soal/penerbitan')
		&& !matchesPathSegment(pathname, '/bank-soal/impor')
		&& !matchesPathSegment(pathname, '/bank-soal/pengaturan');
}

function isSensitiveAssessmentReadPath(pathname: string): boolean {
	const cleanPath = pathname.split('?')[0] ?? pathname;
	return /^\/api\/asesmen\/events\/[^/]+\/results\/?$/.test(cleanPath)
		|| /^\/api\/asesmen\/sessions\/[^/]+\/results\/?$/.test(cleanPath)
		|| /^\/api\/asesmen\/sessions\/[^/]+\/item-analysis\/?$/.test(cleanPath)
		|| /^\/api\/asesmen\/sessions\/[^/]+\/operational-recap\/?$/.test(cleanPath)
		|| /^\/api\/asesmen\/sessions\/[^/]+\/ungraded-essays\/?$/.test(cleanPath)
		|| /^\/api\/asesmen\/sessions\/[^/]+\/score\/?$/.test(cleanPath)
		|| /^\/api\/asesmen\/sessions\/[^/]+\/answers\/[^/]+\/grade-essay\/?$/.test(cleanPath)
		|| /^\/api\/asesmen\/sessions\/[^/]+\/proctoring(?:\/.*)?$/.test(cleanPath)
		|| /^\/api\/asesmen\/sessions\/[^/]+\/rooms\/[^/]+\/proctoring(?:\/.*)?$/.test(cleanPath)
		|| /^\/api\/asesmen\/sessions\/[^/]+\/audit-logs\/?$/.test(cleanPath);
}

function asesmenPermission(pathname: string, method: string): string[] | undefined {
	if (matchesPathSegment(pathname, '/asesmen/aplikasi-siswa/release')) return [];
	if (matchesPathSegment(pathname, '/asesmen/hasil')) return ['asesmen.result_read'];
	if (matchesPathSegment(pathname, '/ujian/command-center') || matchesPathSegment(pathname, '/asesmen/pelaksanaan') || matchesPathSegment(pathname, '/asesmen/pengawasan')) return ['asesmen.proctor'];
	if (/^\/api\/asesmen\/events\/[^/]+\/results\/?$/.test(pathname) || /^\/api\/asesmen\/sessions\/[^/]+\/(results|item-analysis|operational-recap)\/?$/.test(pathname)) {
		return ['asesmen.result_read'];
	}
	if (/^\/api\/asesmen\/sessions\/[^/]+\/(ungraded-essays|score)\/?$/.test(pathname) || /^\/api\/asesmen\/sessions\/[^/]+\/answers\/[^/]+\/grade-essay\/?$/.test(pathname)) {
		return ['asesmen.score'];
	}
	if (/^\/api\/asesmen\/sessions\/[^/]+\/proctoring(?:\/.*)?$/.test(pathname) || /^\/api\/asesmen\/sessions\/[^/]+\/rooms\/[^/]+\/proctoring(?:\/.*)?$/.test(pathname) || /^\/api\/asesmen\/sessions\/[^/]+\/audit-logs\/?$/.test(pathname)) {
		return ['asesmen.proctor'];
	}
	if (matchesPathSegment(pathname, '/asesmen/non-tes') || matchesPathSegment(pathname, '/api/asesmen/non-test-assessments')) {
		return isReadMethod(method) ? ['asesmen.read'] : ['asesmen.score'];
	}
	if (matchesPathSegment(pathname, '/asesmen/aplikasi-siswa')) return ['asesmen.read'];
	if (matchesPathSegment(pathname, '/api/asesmen/proctoring')) return ['asesmen.proctor'];
	if (matchesPathSegment(pathname, '/api/asesmen/approvals')) {
		return [];
	}
	if (matchesPathSegment(pathname, '/asesmen/paket') || matchesPathSegment(pathname, '/api/asesmen/packages')) {
		return isReadMethod(method) ? ['asesmen.read'] : ['asesmen.package_manage'];
	}
	if (matchesPathSegment(pathname, '/asesmen/kegiatan') || matchesPathSegment(pathname, '/api/asesmen/events')) {
		return isReadMethod(method) ? ['asesmen.read'] : ['asesmen.event_manage'];
	}
	if (matchesPathSegment(pathname, '/asesmen/sesi') || matchesPathSegment(pathname, '/api/asesmen/sessions')) {
		return isReadMethod(method) ? ['asesmen.read'] : ['asesmen.proctor'];
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
	if (matchesPathSegment(pathname, '/tu/arsip') || matchesPathSegment(pathname, '/api/tu/archives')) return isReadMethod(method) ? ['archives.read'] : ['archives.manage'];
	if (matchesPathSegment(pathname, '/tu') || matchesPathSegment(pathname, '/api/tu')) return isReadMethod(method) ? ['letters.read'] : ['letters.manage'];
	return undefined;
}

function schedulePermission(pathname: string): string[] | undefined {
	if (!SCHEDULE_PAGE_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix))) return undefined;
	return ['academic.read', 'student_portal.schedule_read', 'parent_portal.child_schedule_read'];
}

function gradesPermission(pathname: string, method: string): string[] | undefined {
	if (!GRADES_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix))) return undefined;
	return isReadMethod(method) ? ['grades.read', 'grades.manage'] : ['grades.manage'];
}

function journalPermission(pathname: string, method: string): string[] | undefined {
	if (!JOURNAL_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix))) return undefined;
	return isReadMethod(method)
		? ['journal.read', 'journal.manage', 'journal.read_all', 'journal.manage_all']
		: ['journal.manage', 'journal.manage_all'];
}

function employeePermission(pathname: string, method: string): string[] | undefined {
	if (!EMPLOYEE_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix))) return undefined;
	return isReadMethod(method) ? ['employees.read', 'employees.manage'] : ['employees.manage'];
}

function systemBackupPermission(pathname: string, method: string): string[] | undefined {
	if (!matchesPathSegment(pathname, '/api/system/backups')) return undefined;
	if (method === 'POST' && /^\/api\/system\/backups\/[^/]+\/(validate-restore|restore-command)\/?$/.test(pathname)) return ['backup.restore_plan'];
	if (method === 'POST' && pathname === '/api/system/backups/run') return ['backup.create'];
	if (!isReadMethod(method)) return [];
	if (/^\/api\/system\/backups\/[^/]+\/download\/?$/.test(pathname)) return ['backup.download'];
	return ['backup.read'];
}

function isRombelTimetableJournalSessionPath(pathname: string): boolean {
	return /^\/api\/academic\/rombel\/[^/]+\/timetable-slots\/[^/]+\/journal-session\/?$/.test(pathname.split('?')[0] ?? pathname);
}

export function requiredPermissionsForPath(pathname: string, method: string): string[] {
	const settings = settingsPermission(pathname);
	if (settings) return settings;
	if (matchesPathSegment(pathname, '/api/internal-analytics/events')) return [];
	const systemBackups = systemBackupPermission(pathname, method);
	if (systemBackups) return systemBackups;
	if (matchesPathSegment(pathname, '/api/internal-analytics/export')) return ['analytics.export'];
	if (matchesPathSegment(pathname, '/api/internal-analytics')) return isReadMethod(method) ? ['analytics.read'] : ['analytics.read'];
	if (matchesPathSegment(pathname, '/api/rbac')) return isReadMethod(method) ? ['roles.read'] : ['roles.manage'];
	if (matchesPathSegment(pathname, '/parents') || matchesPathSegment(pathname, '/api/parents')) return isReadMethod(method) ? ['parents.read'] : ['parents.manage'];
	if (isRombelTimetableJournalSessionPath(pathname)) return ['journal.manage', 'journal.manage_all'];
	if (matchesPathSegment(pathname, '/academic') || matchesPathSegment(pathname, '/akademik') || matchesPathSegment(pathname, '/api/academic')) return isReadMethod(method) ? ['academic.read'] : ['academic.manage'];
	if (matchesPathSegment(pathname, '/pusaka') || matchesPathSegment(pathname, '/api/pusaka')) return isReadMethod(method) ? ['pusaka.read'] : ['pusaka.manage'];
	if (matchesPathSegment(pathname, '/website') || matchesPathSegment(pathname, '/api/website')) return isReadMethod(method) ? ['website.read'] : ['website.manage'];
	if (matchesPathSegment(pathname, '/notifications') || matchesPathSegment(pathname, '/api/notifications')) return ['notifications.read'];
	if (matchesPathSegment(pathname, '/api/portal/guru/timetable')) return ['academic.read', 'journal.read', 'journal.manage', 'journal.read_all', 'journal.manage_all'];
	if (isStudentPortalPath(pathname)) return ['student_portal.read'];
	if (isParentPortalPath(pathname)) return ['parent_portal.read'];
	if (matchesPathSegment(pathname, '/kesiswaan/kartu-siswa/scan') || matchesPathSegment(pathname, '/api/kesiswaan/kartu-siswa/scan')) return ['id_cards.scan', 'id_cards.manage'];
	if (matchesPathSegment(pathname, '/kesiswaan/kartu-siswa') || matchesPathSegment(pathname, '/api/kesiswaan/kartu-siswa')) return isReadMethod(method) ? ['id_cards.read', 'id_cards.manage'] : ['id_cards.manage'];
	if (isKesiswaanPath(pathname)) return isReadMethod(method) ? ['kesiswaan.read', 'students.read'] : ['kesiswaan.manage', 'students.manage'];
	if (isStudentPagePath(pathname) || isStudentApiPath(pathname)) return isReadMethod(method) ? ['students.read'] : ['students.manage'];
	return usersPermission(pathname, method)
		?? schedulePermission(pathname)
		?? gradesPermission(pathname, method)
		?? journalPermission(pathname, method)
		?? employeePermission(pathname, method)
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
	if (
		isGuruSafeAssessmentSupportReadPath(pathname, method)
		&& hasAnyPermission(user, [
			'bank_soal.read',
			'bank_soal.create',
			'bank_soal.update',
			'bank_soal.review',
			'bank_soal.publish',
			'bank_soal.import',
			'bank_soal.analytics',
			'bank_soal.update_own',
			'bank_soal.submit',
			'bank_soal.approve',
			'bank_soal.read_all',
			'bank_soal.use_in_package'
		])
	) return true;

	if (isAdminOnlyPath(pathname) && !isGuruSafeAssessmentSupportReadPath(pathname, method)) return false;
	if (isStudentPortalPath(pathname)) return hasAnyRole(user, ['siswa']);
	if (isParentPortalPath(pathname)) return hasAnyRole(user, ['ortu']);

	// Default-authenticated surfaces such as /settings/account remain available,
	// but paths with explicit permission metadata must fail closed when the user
	// lacks both dynamic permission and legacy fallback role.
	return requiredPermissions.length === 0;
}
