import { describe, expect, it } from 'vitest';

import { canAccessProtectedRoute, hasAnyPermission, hasAnyRole, isAdminOnlyPath, isBankSoalPath, isKesiswaanPath, isMustChangePasswordAllowedPath, isPublicPath, isReadMethod, isStaffOperationPath, isStudentApiPath, isStudentPagePath, requiredPermissionsForPath } from './route-access';

describe('route access helpers', () => {
	it('keeps account settings available while gating system settings', () => {
		expect(isAdminOnlyPath('/settings')).toBe(false);
		expect(isAdminOnlyPath('/settings/account')).toBe(false);
		expect(isAdminOnlyPath('/api/auth/account')).toBe(false);
		expect(isAdminOnlyPath('/settings/users')).toBe(true);
		expect(isAdminOnlyPath('/settings/rbac')).toBe(true);
		expect(isAdminOnlyPath('/settings/user-change-requests')).toBe(true);
		expect(isAdminOnlyPath('/settings/backups')).toBe(true);
	});

	it('matches public paths and prefixes', () => {
		expect(isPublicPath('/')).toBe(true);
		expect(isPublicPath('/berita/arsip-kegiatan')).toBe(true);
		expect(isPublicPath('/api/cbt-portal/participants/participant-1/start')).toBe(false);
		expect(isPublicPath('/ujian')).toBe(false);
		expect(isPublicPath('/pengawas-ujian')).toBe(false);
		expect(isPublicPath('/beritaship')).toBe(false);
		expect(isPublicPath('/dashboard')).toBe(false);
	});

	it('matches staff and kesiswaan operation paths', () => {
		expect(isStaffOperationPath('/library/loans')).toBe(true);
		expect(isStaffOperationPath('/api/inventory/items')).toBe(true);
		expect(isStaffOperationPath('/parents')).toBe(false);
		expect(isStaffOperationPath('/parentship')).toBe(false);
		expect(isKesiswaanPath('/kesiswaan')).toBe(true);
		expect(isKesiswaanPath('/api/kesiswaan/students')).toBe(true);
		expect(isKesiswaanPath('/kesiswaanship')).toBe(false);
	});

	it('matches admin-only paths on exact segment boundaries', () => {
		expect(isAdminOnlyPath('/parents')).toBe(true);
		expect(isAdminOnlyPath('/parents/parent-1')).toBe(true);
		expect(isAdminOnlyPath('/api/parents')).toBe(true);
		expect(isAdminOnlyPath('/api/parents/parent-1')).toBe(true);
		expect(isAdminOnlyPath('/api/bank-soal/questions')).toBe(false);
		expect(isAdminOnlyPath('/bank-soal')).toBe(false);
		expect(isAdminOnlyPath('/bank-soal/tambah')).toBe(false);
		expect(isAdminOnlyPath('/bank-soal/impor')).toBe(false);
		expect(isAdminOnlyPath('/bank-soal/verifikasi')).toBe(false);
		expect(isAdminOnlyPath('/api/school-profile')).toBe(true);
		expect(isAdminOnlyPath('/parentship')).toBe(false);
		expect(isAdminOnlyPath('/api/parentship')).toBe(false);
	});

	it('keeps Bank Soal route boundaries explicit for admin and guru access', () => {
		expect(isPublicPath('/bank-soal')).toBe(false);
		expect(isAdminOnlyPath('/bank-soal')).toBe(false);
		const finalBankSoalRoutes = [
			'/bank-soal',
			'/bank-soal/daftar',
			'/bank-soal/tambah',
			'/bank-soal/verifikasi',
			'/bank-soal/impor',
			'/bank-soal/alat',
			'/bank-soal/analisis-butir',
			'/bank-soal/mapel-kd',
			'/bank-soal/pengaturan'
		];
		expect(finalBankSoalRoutes.map((route) => [route, isAdminOnlyPath(route)])).toEqual(
			finalBankSoalRoutes.map((route) => [route, false])
		);
		expect(isAdminOnlyPath('/api/bank-soal/questions')).toBe(false);
		expect(isAdminOnlyPath('/api/bank-soal/questions/export')).toBe(false);
		expect(isAdminOnlyPath('/api/bank-soal/soal-support/subjects')).toBe(false);
		expect(isBankSoalPath('/bank-soal/analisis-butir')).toBe(true);
		expect(isBankSoalPath('/api/bank-soal/summary')).toBe(true);
		expect(isBankSoalPath('/bank-soalship')).toBe(false);
	});




	it('matches student master paths without broad admin-only exposure for read BFF', () => {
		expect(isStudentPagePath('/students')).toBe(true);
		expect(isStudentPagePath('/studentship')).toBe(false);
		expect(isStudentApiPath('/api/students')).toBe(true);
		expect(isStudentApiPath('/api/students/student-1')).toBe(true);
		expect(isStudentApiPath('/api/studentship')).toBe(false);
		expect(isAdminOnlyPath('/api/students')).toBe(false);
	});

	it('classifies read methods for method-aware BFF guards', () => {
		expect(isReadMethod('GET')).toBe(true);
		expect(isReadMethod('HEAD')).toBe(true);
		expect(isReadMethod('OPTIONS')).toBe(true);
		expect(isReadMethod('POST')).toBe(false);
		expect(isReadMethod('PATCH')).toBe(false);
	});



	it('allows protected routes by dynamic permissions before legacy role fallback', () => {
		const user = { id: '1', username: 'operator', role: '', roles: [], permissions: ['users.read', 'bank_soal.read', 'profile_changes.review'] };

		expect(canAccessProtectedRoute(user, '/settings/users', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...user, permissions: [...user.permissions, 'roles.read'] }, '/settings/rbac', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(user, '/settings/user-change-requests', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/settings/account', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/settings', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ ...user, permissions: ['settings.account'] }, '/settings', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/api/auth/account', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/api/auth/account/change-request-fields', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(user, '/bank-soal/daftar', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/settings/users', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/settings/rbac', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ id: '6', username: 'rbac-reader', role: '', roles: [], permissions: ['roles.read'] }, '/settings/rbac', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ id: '7', username: 'plain-guru', role: 'guru', roles: ['guru'], permissions: [] }, '/settings/rbac', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ id: '8', username: 'rbac-mutator', role: '', roles: [], permissions: ['roles.manage'] }, '/api/rbac/roles/guru/permissions', 'PUT')).toBe(true);
		expect(canAccessProtectedRoute({ id: '9', username: 'rbac-reader', role: '', roles: [], permissions: ['roles.read'] }, '/api/rbac/roles/guru/permissions', 'PUT')).toBe(false);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/settings/user-change-requests', 'GET')).toBe(false);
	});

	it('keeps internal analytics routes authenticated and read dashboard permission-scoped', () => {
		const plainUser = { id: '1', username: 'plain', role: 'guru', roles: ['guru'], permissions: [] };
		const analyticsReader = { id: '2', username: 'reader', role: '', roles: [], permissions: ['analytics.read'] };
		const analyticsExporter = { id: '3', username: 'exporter', role: '', roles: [], permissions: ['analytics.export'] };

		expect(isPublicPath('/api/public/analytics/events')).toBe(true);
		expect(isPublicPath('/api/internal-analytics/events')).toBe(false);
		expect(requiredPermissionsForPath('/api/internal-analytics/events', 'POST')).toEqual([]);
		expect(canAccessProtectedRoute(undefined, '/api/internal-analytics/events', 'POST')).toBe(false);
		expect(canAccessProtectedRoute(plainUser, '/api/internal-analytics/events', 'POST')).toBe(true);

		expect(requiredPermissionsForPath('/settings/analytics', 'GET')).toEqual(['analytics.read']);
		expect(requiredPermissionsForPath('/api/internal-analytics/summary', 'GET')).toEqual(['analytics.read']);
		expect(requiredPermissionsForPath('/api/internal-analytics/daily', 'GET')).toEqual(['analytics.read']);
		expect(requiredPermissionsForPath('/api/internal-analytics/export', 'GET')).toEqual(['analytics.export']);
		expect(requiredPermissionsForPath('/settings/backups', 'GET')).toEqual(['backup.read']);
		expect(requiredPermissionsForPath('/api/system/backups/status', 'GET')).toEqual(['backup.read']);
	expect(requiredPermissionsForPath('/api/system/backups', 'GET')).toEqual(['backup.read']);
	expect(requiredPermissionsForPath('/api/system/backups/offsite', 'GET')).toEqual(['backup.read']);
	expect(requiredPermissionsForPath('/api/system/backups/run', 'POST')).toEqual(['backup.create']);
	expect(requiredPermissionsForPath('/api/system/backups/jobs/backup-20260513-110000', 'GET')).toEqual(['backup.read']);
	expect(requiredPermissionsForPath('/api/system/backups/pusaka_20260513_000001.dump/validate-restore', 'POST')).toEqual(['backup.restore_plan']);
	expect(requiredPermissionsForPath('/api/system/backups/pusaka_20260513_000001.dump/restore-command', 'POST')).toEqual(['backup.restore_plan']);
	expect(requiredPermissionsForPath('/api/system/backups/pusaka_20260513_000001.dump/download', 'GET')).toEqual(['backup.download']);
		expect(canAccessProtectedRoute(plainUser, '/settings/analytics', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(analyticsReader, '/settings/analytics', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(analyticsReader, '/api/internal-analytics/summary', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(analyticsReader, '/api/internal-analytics/export', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(analyticsExporter, '/api/internal-analytics/export', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(plainUser, '/settings/backups', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ ...plainUser, permissions: ['backup.read'] }, '/settings/backups', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...plainUser, permissions: ['backup.read'] }, '/api/system/backups/status', 'GET')).toBe(true);
	expect(canAccessProtectedRoute({ ...plainUser, permissions: ['backup.read'] }, '/api/system/backups/offsite', 'GET')).toBe(true);
	expect(canAccessProtectedRoute({ ...plainUser, permissions: ['backup.read'] }, '/api/system/backups/run', 'POST')).toBe(false);
	expect(canAccessProtectedRoute({ ...plainUser, permissions: ['backup.create'] }, '/api/system/backups/run', 'POST')).toBe(true);
	expect(canAccessProtectedRoute({ ...plainUser, permissions: ['backup.read'] }, '/api/system/backups/pusaka_20260513_000001.dump/validate-restore', 'POST')).toBe(false);
	expect(canAccessProtectedRoute({ ...plainUser, permissions: ['backup.restore_plan'] }, '/api/system/backups/pusaka_20260513_000001.dump/validate-restore', 'POST')).toBe(true);
	expect(canAccessProtectedRoute({ ...plainUser, permissions: ['backup.read'] }, '/api/system/backups/pusaka_20260513_000001.dump/download', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ ...plainUser, permissions: ['backup.download'] }, '/api/system/backups/pusaka_20260513_000001.dump/download', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ id: '4', username: 'admin', role: 'admin', roles: ['admin'], permissions: [] }, '/settings/backups', 'GET')).toBe(true);
	});

	it('guards student and parent portal routes by dedicated portal roles or permissions', () => {
		const siswa = { id: '1', username: 'siswa', role: 'siswa', roles: ['siswa'], permissions: [] };
		const ortu = { id: '2', username: 'ortu', role: 'ortu', roles: ['ortu'], permissions: [] };
		const studentPermission = { id: '3', username: 'student-perm', role: '', roles: [], permissions: ['student_portal.read'] };
		const parentPermission = { id: '4', username: 'parent-perm', role: '', roles: [], permissions: ['parent_portal.read'] };

		expect(canAccessProtectedRoute(undefined, '/portal/siswa', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(siswa, '/portal/siswa', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(studentPermission, '/portal/siswa', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(ortu, '/portal/siswa', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ ...ortu, roles: ['admin'] }, '/portal/siswa', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(ortu, '/portal/orang-tua', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(parentPermission, '/portal/orang-tua', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(parentPermission, '/api/portal/parent/me', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(siswa, '/portal/orang-tua', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(siswa, '/api/portal/parent/me', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(siswa, '/api/portal/siswa/profile', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(siswa, '/api/portal/student/me', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(studentPermission, '/api/portal/student/me', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(ortu, '/api/portal/siswa/profile', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(ortu, '/api/portal/student/me', 'GET')).toBe(false);
	});

	it('limits first-login users to password change and logout surfaces', () => {
		const mustChangeUser = {
			id: '1',
			username: 'siswa',
			role: 'siswa',
			roles: ['siswa'],
			permissions: ['student_portal.read'],
			must_change_password: true
		};

		expect(isMustChangePasswordAllowedPath('/settings/account', 'GET')).toBe(true);
		expect(isMustChangePasswordAllowedPath('/api/auth/account', 'GET')).toBe(true);
		expect(isMustChangePasswordAllowedPath('/api/auth/change-password', 'POST')).toBe(true);
		expect(isMustChangePasswordAllowedPath('/api/auth/logout', 'POST')).toBe(true);
		expect(isMustChangePasswordAllowedPath('/api/auth/preferences/sidebar', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(mustChangeUser, '/settings/account', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(mustChangeUser, '/portal/siswa', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ ...mustChangeUser, must_change_password: false }, '/portal/siswa', 'GET')).toBe(true);
	});


	it('keeps mutation route checks permission-specific', () => {
		const reader = { id: '1', username: 'reader', role: '', roles: [], permissions: ['users.read', 'bank_soal.read'] };
		const mutator = { id: '2', username: 'mutator', role: '', roles: [], permissions: ['users.create', 'bank_soal.create'] };
		const profileReviewer = { id: '3', username: 'reviewer', role: '', roles: [], permissions: ['profile_changes.review'] };
		const plainGuru = { id: '4', username: 'guru', role: 'guru', roles: ['guru'], permissions: [] };
		const bankSoalEditor = { id: '5', username: 'editor', role: '', roles: [], permissions: ['bank_soal.update'] };
		const bankSoalReviewer = { id: '6', username: 'reviewer', role: '', roles: [], permissions: ['bank_soal.review'] };

		expect(canAccessProtectedRoute(reader, '/api/users', 'POST')).toBe(false);
		expect(canAccessProtectedRoute(mutator, '/api/users', 'POST')).toBe(true);
		expect(canAccessProtectedRoute(reader, '/api/users/change-requests/request-1', 'PATCH')).toBe(false);
		expect(canAccessProtectedRoute(profileReviewer, '/api/users/change-requests/request-1', 'PATCH')).toBe(true);
		expect(canAccessProtectedRoute(profileReviewer, '/api/users/change-requests/pending-count', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(profileReviewer, '/api/users/change-requests/export', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(reader, '/api/bank-soal/questions', 'POST')).toBe(false);
		expect(canAccessProtectedRoute(mutator, '/api/bank-soal/questions', 'POST')).toBe(true);
		expect(canAccessProtectedRoute(plainGuru, '/api/bank-soal/questions', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(plainGuru, '/api/bank-soal/questions', 'POST')).toBe(false);
		expect(canAccessProtectedRoute(plainGuru, '/api/bank-soal/questions/question-1', 'PUT')).toBe(false);
		expect(canAccessProtectedRoute(plainGuru, '/api/bank-soal/questions/question-1/workflow', 'PATCH')).toBe(false);
		expect(canAccessProtectedRoute(bankSoalEditor, '/api/bank-soal/questions/question-1', 'PUT')).toBe(true);
		expect(canAccessProtectedRoute(bankSoalEditor, '/api/bank-soal/questions/question-1/workflow', 'PATCH')).toBe(false);
		expect(canAccessProtectedRoute(bankSoalReviewer, '/api/bank-soal/questions/question-1/workflow', 'PATCH')).toBe(true);
		expect(canAccessProtectedRoute(plainGuru, '/api/bank-soal/questions/question-1', 'DELETE')).toBe(false);
		expect(canAccessProtectedRoute(plainGuru, '/bank-soal/daftar', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ ...plainGuru, permissions: ['bank_soal.read'] }, '/bank-soal/daftar', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(plainGuru, '/bank-soal/tambah', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(mutator, '/bank-soal/tambah', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(plainGuru, '/bank-soal/verifikasi', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(plainGuru, '/bank-soal/impor', 'GET')).toBe(false);
		expect(canAccessProtectedRoute(plainGuru, '/bank-soal/pengaturan', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ id: '4', username: 'guru', role: 'guru', roles: ['guru'], permissions: [] }, '/api/academic/rombel/class-1/timetable-slots/slot-1/journal-session', 'POST')).toBe(false);
		expect(canAccessProtectedRoute({ id: '5', username: 'journal-all', role: '', roles: [], permissions: ['journal.manage_all'] }, '/api/academic/rombel/class-1/timetable-slots/slot-1/journal-session', 'POST')).toBe(true);
		expect(canAccessProtectedRoute(reader, '/api/academic/rombel/class-1/timetable-slots/slot-1/journal-session', 'POST')).toBe(false);
		expect(canAccessProtectedRoute({ ...plainGuru, permissions: ['academic.read'] }, '/jadwal', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(plainGuru, '/jadwal', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ ...plainGuru, permissions: ['grades.read'] }, '/grades', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(plainGuru, '/grades', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ ...plainGuru, permissions: ['journal.manage'] }, '/journal', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(plainGuru, '/journal', 'GET')).toBe(false);
	});

	it('documents route permission requirements for main migrated modules', () => {
		expect(requiredPermissionsForPath('/settings/users', 'GET')).toEqual(['users.read']);
		expect(requiredPermissionsForPath('/settings', 'GET')).toEqual(['settings.account']);
		expect(requiredPermissionsForPath('/settings/rbac', 'GET')).toEqual(['roles.read']);
		expect(requiredPermissionsForPath('/settings/analytics', 'GET')).toEqual(['analytics.read']);
		expect(requiredPermissionsForPath('/notifications', 'GET')).toEqual(['notifications.read']);
		expect(requiredPermissionsForPath('/api/notifications', 'GET')).toEqual(['notifications.read']);
		expect(requiredPermissionsForPath('/jadwal', 'GET')).toEqual(['academic.read', 'student_portal.schedule_read', 'parent_portal.child_schedule_read']);
		expect(requiredPermissionsForPath('/grades', 'GET')).toEqual(['grades.read', 'grades.manage']);
		expect(requiredPermissionsForPath('/api/grades/components', 'POST')).toEqual(['grades.manage']);
		expect(requiredPermissionsForPath('/journal', 'GET')).toEqual(['journal.read', 'journal.manage', 'journal.read_all', 'journal.manage_all']);
		expect(requiredPermissionsForPath('/api/journal/sessions', 'POST')).toEqual(['journal.manage', 'journal.manage_all']);
		expect(requiredPermissionsForPath('/employees', 'GET')).toEqual(['employees.read', 'employees.manage']);
		expect(requiredPermissionsForPath('/api/employees/employee-1', 'PUT')).toEqual(['employees.manage']);
		expect(requiredPermissionsForPath('/api/rbac/matrix', 'GET')).toEqual(['roles.read']);
		expect(requiredPermissionsForPath('/api/internal-analytics/daily', 'GET')).toEqual(['analytics.read']);
		expect(requiredPermissionsForPath('/api/internal-analytics/export', 'GET')).toEqual(['analytics.export']);
		expect(requiredPermissionsForPath('/api/rbac/roles/guru/permissions', 'PUT')).toEqual(['roles.manage']);
		expect(requiredPermissionsForPath('/settings/user-change-requests', 'GET')).toEqual(['profile_changes.review']);
		expect(requiredPermissionsForPath('/api/users', 'POST')).toEqual(['users.create']);
		expect(requiredPermissionsForPath('/api/users/student-accounts/preview', 'GET')).toEqual(['student_accounts.manage']);
		expect(requiredPermissionsForPath('/api/users/student-accounts/generate', 'POST')).toEqual(['student_accounts.manage']);
		expect(requiredPermissionsForPath('/api/users/parent-accounts/preview', 'GET')).toEqual(['parent_accounts.manage']);
		expect(requiredPermissionsForPath('/api/users/parent-accounts/generate', 'POST')).toEqual(['parent_accounts.manage']);
		expect(requiredPermissionsForPath('/api/users/change-requests', 'GET')).toEqual(['profile_changes.review']);
		expect(requiredPermissionsForPath('/api/users/change-requests/pending-count', 'GET')).toEqual(['profile_changes.review']);
		expect(requiredPermissionsForPath('/api/users/change-requests/export', 'GET')).toEqual(['profile_changes.review']);
		expect(requiredPermissionsForPath('/api/users/change-requests/request-1', 'PATCH')).toEqual(['profile_changes.review']);
		expect(requiredPermissionsForPath('/api/users/user-1/reset-password', 'POST')).toEqual(['users.reset_password']);
		expect(requiredPermissionsForPath('/api/users/user-1/force-password-change', 'POST')).toEqual(['users.reset_password']);
		expect(requiredPermissionsForPath('/api/users/user-1/status', 'PATCH')).toEqual(['users.deactivate']);
		expect(requiredPermissionsForPath('/api/users/user-1/profile-link', 'PATCH')).toEqual(['users.update']);
		expect(requiredPermissionsForPath('/bank-soal/tambah', 'GET')).toEqual(['bank_soal.create']);
		expect(requiredPermissionsForPath('/bank-soal/verifikasi', 'GET')).toEqual(['bank_soal.review']);
		expect(requiredPermissionsForPath('/bank-soal/impor', 'GET')).toEqual(['bank_soal.import']);
		expect(requiredPermissionsForPath('/bank-soal/alat', 'GET')).toEqual(['bank_soal.analytics', 'bank_soal.publish', 'bank_soal.settings']);
		expect(requiredPermissionsForPath('/bank-soal/pengaturan', 'GET')).toEqual(['bank_soal.settings']);
		expect(requiredPermissionsForPath('/api/bank-soal/questions', 'POST')).toEqual(['bank_soal.create']);
		expect(requiredPermissionsForPath('/api/bank-soal/questions/import-legacy', 'POST')).toEqual(['bank_soal.import']);
		expect(requiredPermissionsForPath('/api/bank-soal/questions/bulk-workflow', 'PATCH')).toEqual(['bank_soal.review', 'bank_soal.publish']);
		expect(requiredPermissionsForPath('/api/bank-soal/questions/question-1/workflow', 'PATCH')).toEqual(['bank_soal.review', 'bank_soal.publish']);
		expect(requiredPermissionsForPath('/api/bank-soal/questions/question-1', 'PUT')).toEqual(['bank_soal.update']);
		expect(requiredPermissionsForPath('/api/bank-soal/questions/question-1/duplicate', 'POST')).toEqual(['bank_soal.create']);
		expect(requiredPermissionsForPath('/api/bank-soal/questions/question-1/revision', 'POST')).toEqual(['bank_soal.update']);
		expect(requiredPermissionsForPath('/api/bank-soal/questions/question-1', 'DELETE')).toEqual(['bank_soal.delete']);
		expect(requiredPermissionsForPath('/api/bank-soal/assets', 'POST')).toEqual(['bank_soal.create', 'bank_soal.update']);
		expect(requiredPermissionsForPath('/api/academic/rombel/class-1/timetable-slots/slot-1/journal-session', 'POST')).toEqual(['journal.manage', 'journal.manage_all']);
		expect(requiredPermissionsForPath('/portal/siswa', 'GET')).toEqual(['student_portal.read']);
		expect(requiredPermissionsForPath('/api/portal/siswa/results', 'GET')).toEqual(['student_portal.read']);
		expect(requiredPermissionsForPath('/api/portal/student/me', 'GET')).toEqual(['student_portal.read']);
		expect(requiredPermissionsForPath('/portal/orang-tua', 'GET')).toEqual(['parent_portal.read']);
		expect(requiredPermissionsForPath('/api/portal/orang-tua/children', 'GET')).toEqual(['parent_portal.read']);
		expect(requiredPermissionsForPath('/api/portal/parent/me', 'GET')).toEqual(['parent_portal.read']);
		expect(requiredPermissionsForPath('/tu/arsip', 'GET')).toEqual(['archives.read']);
		expect(requiredPermissionsForPath('/api/tu/archives/documents', 'POST')).toEqual(['archives.manage']);
	});

	it('checks role membership from user payloads', () => {
		expect(hasAnyRole({ id: '1', username: 'guru', role: 'guru', roles: [], permissions: [] }, ['guru'])).toBe(true);
		expect(hasAnyRole({ id: '1', username: 'staf', role: '', roles: ['guru', 'staf'], permissions: [] }, ['admin', 'staf'])).toBe(true);
		expect(hasAnyRole({ id: '1', username: 'ortu', role: 'ortu', roles: [], permissions: [] }, ['admin', 'guru'])).toBe(false);
		expect(hasAnyPermission({ id: '1', username: 'rbac', role: '', roles: [], permissions: [' users.read '] }, ['users.read'])).toBe(true);
	});
});
