import { describe, expect, it } from 'vitest';

import { canAccessProtectedRoute, hasAnyPermission, hasAnyRole, isAdminOnlyPath, isBankSoalPath, isGuruSafeAssessmentSupportReadPath, isKesiswaanPath, isPublicPath, isReadMethod, isStaffOperationPath, isStudentApiPath, isStudentPagePath, requiredPermissionsForPath } from './route-access';

describe('route access helpers', () => {
	it('keeps settings root available to authenticated non-admin users', () => {
		expect(isAdminOnlyPath('/settings')).toBe(false);
		expect(isAdminOnlyPath('/settings/account')).toBe(false);
		expect(isAdminOnlyPath('/api/auth/account')).toBe(false);
		expect(isAdminOnlyPath('/settings/users')).toBe(true);
		expect(isAdminOnlyPath('/settings/user-change-requests')).toBe(true);
	});

	it('matches public paths and prefixes', () => {
		expect(isPublicPath('/')).toBe(true);
		expect(isPublicPath('/berita/arsip-kegiatan')).toBe(true);
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
		expect(isAdminOnlyPath('/asesmen/kegiatan')).toBe(true);
		expect(isAdminOnlyPath('/asesmen/kegiatan/event-1')).toBe(true);
		expect(isAdminOnlyPath('/api/asesmen/events')).toBe(true);
		expect(isAdminOnlyPath('/api/asesmen/events/event-1')).toBe(true);
		expect(isAdminOnlyPath('/asesmen/paket')).toBe(true);
		expect(isAdminOnlyPath('/asesmen/sesi')).toBe(true);
		expect(isAdminOnlyPath('/api/asesmen/packages')).toBe(true);
		expect(isAdminOnlyPath('/api/asesmen/sessions/session-1')).toBe(true);
		expect(isAdminOnlyPath('/api/bank-soal/questions')).toBe(false);
		expect(isAdminOnlyPath('/bank-soal')).toBe(false);
		expect(isAdminOnlyPath('/bank-soal/tambah')).toBe(false);
		expect(isAdminOnlyPath('/bank-soal/impor')).toBe(false);
		expect(isAdminOnlyPath('/bank-soal/verifikasi')).toBe(false);
		expect(isAdminOnlyPath('/asesmen/kegiatan/event-1/members')).toBe(true);
		expect(isAdminOnlyPath('/api/school-profile')).toBe(true);
		expect(isAdminOnlyPath('/parentship')).toBe(false);
		expect(isAdminOnlyPath('/api/parentship')).toBe(false);
	});

	it('keeps Bank Soal and Asesmen route boundaries explicit for admin and guru access', () => {
		expect(isPublicPath('/bank-soal')).toBe(false);
		expect(isAdminOnlyPath('/bank-soal')).toBe(false);
		const finalBankSoalRoutes = [
			'/bank-soal',
			'/bank-soal/daftar',
			'/bank-soal/tambah',
			'/bank-soal/verifikasi',
			'/bank-soal/impor',
			'/bank-soal/analisis-butir',
			'/bank-soal/mapel-kd',
			'/bank-soal/pengaturan'
		];
		expect(finalBankSoalRoutes.map((route) => [route, isAdminOnlyPath(route)])).toEqual(
			finalBankSoalRoutes.map((route) => [route, false])
		);
		expect(isAdminOnlyPath('/api/bank-soal/questions')).toBe(false);
		expect(isAdminOnlyPath('/api/bank-soal/questions/export')).toBe(false);
		expect(isAdminOnlyPath('/asesmen/kegiatan')).toBe(true);
		expect(isAdminOnlyPath('/asesmen/kegiatan/event-1/exam-cards')).toBe(true);
		expect(isAdminOnlyPath('/api/asesmen/events/event-1')).toBe(true);
		expect(isAdminOnlyPath('/api/bank-soal/soal-support/subjects')).toBe(false);
		expect(isBankSoalPath('/bank-soal/analisis-butir')).toBe(true);
		expect(isBankSoalPath('/api/bank-soal/summary')).toBe(true);
		expect(isBankSoalPath('/bank-soalship')).toBe(false);
		expect(isGuruSafeAssessmentSupportReadPath('/api/asesmen/events', 'GET')).toBe(true);
		expect(isGuruSafeAssessmentSupportReadPath('/api/asesmen/events', 'POST')).toBe(false);
		expect(isGuruSafeAssessmentSupportReadPath('/api/asesmen/events/event-1', 'GET')).toBe(true);
		expect(isGuruSafeAssessmentSupportReadPath('/api/bank-soal/soal-support/subjects', 'GET')).toBe(true);
		expect(isGuruSafeAssessmentSupportReadPath('/api/asesmen/events/event-1/question-targets', 'GET')).toBe(true);
		expect(isGuruSafeAssessmentSupportReadPath('/api/asesmen/packages?event_id=event-1', 'GET')).toBe(false);
		expect(isGuruSafeAssessmentSupportReadPath('/api/asesmen/packages', 'GET')).toBe(true);
		expect(isGuruSafeAssessmentSupportReadPath('/api/asesmen/sessions/session-1', 'GET')).toBe(true);
		expect(isGuruSafeAssessmentSupportReadPath('/asesmen/kegiatan/event-1', 'GET')).toBe(true);
		expect(isGuruSafeAssessmentSupportReadPath('/api/asesmen/events/event-1/question-targets', 'PUT')).toBe(false);
		expect(isAdminOnlyPath('/asesmen/paket')).toBe(true);
		expect(isAdminOnlyPath('/asesmen/sesi')).toBe(true);
		expect(isAdminOnlyPath('/asesmen/kegiatanhip')).toBe(false);
		expect(isAdminOnlyPath('/api/asesmen/eventship')).toBe(false);
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
		const user = { id: '1', username: 'operator', role: '', roles: [], permissions: ['users.read', 'bank_soal.read', 'asesmen.read', 'profile_changes.review'] };

		expect(canAccessProtectedRoute(user, '/settings/users', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(user, '/settings/user-change-requests', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/settings/account', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/api/auth/account', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/api/auth/account/change-request-fields', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(user, '/bank-soal/daftar', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(user, '/asesmen/kegiatan', 'GET')).toBe(true);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/settings/users', 'GET')).toBe(false);
		expect(canAccessProtectedRoute({ ...user, permissions: [] }, '/settings/user-change-requests', 'GET')).toBe(false);
	});

	it('keeps mutation route checks permission-specific', () => {
		const reader = { id: '1', username: 'reader', role: '', roles: [], permissions: ['users.read', 'bank_soal.read', 'asesmen.read'] };
		const mutator = { id: '2', username: 'mutator', role: '', roles: [], permissions: ['users.create', 'bank_soal.create', 'asesmen.event_manage'] };
		const profileReviewer = { id: '3', username: 'reviewer', role: '', roles: [], permissions: ['profile_changes.review'] };

		expect(canAccessProtectedRoute(reader, '/api/users', 'POST')).toBe(false);
		expect(canAccessProtectedRoute(mutator, '/api/users', 'POST')).toBe(true);
		expect(canAccessProtectedRoute(reader, '/api/users/change-requests/request-1', 'PATCH')).toBe(false);
		expect(canAccessProtectedRoute(profileReviewer, '/api/users/change-requests/request-1', 'PATCH')).toBe(true);
		expect(canAccessProtectedRoute(profileReviewer, '/api/users/change-requests/pending-count', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(profileReviewer, '/api/users/change-requests/export', 'GET')).toBe(true);
		expect(canAccessProtectedRoute(reader, '/api/bank-soal/questions', 'POST')).toBe(false);
		expect(canAccessProtectedRoute(mutator, '/api/bank-soal/questions', 'POST')).toBe(true);
		expect(canAccessProtectedRoute(reader, '/api/asesmen/events/event-1/question-targets', 'PUT')).toBe(false);
		expect(canAccessProtectedRoute(mutator, '/api/asesmen/events/event-1/question-targets', 'PUT')).toBe(true);
	});

	it('documents route permission requirements for main migrated modules', () => {
		expect(requiredPermissionsForPath('/settings/users', 'GET')).toEqual(['users.read']);
		expect(requiredPermissionsForPath('/settings/user-change-requests', 'GET')).toEqual(['profile_changes.review']);
		expect(requiredPermissionsForPath('/api/users', 'POST')).toEqual(['users.create']);
		expect(requiredPermissionsForPath('/api/users/change-requests', 'GET')).toEqual(['profile_changes.review']);
		expect(requiredPermissionsForPath('/api/users/change-requests/pending-count', 'GET')).toEqual(['profile_changes.review']);
		expect(requiredPermissionsForPath('/api/users/change-requests/export', 'GET')).toEqual(['profile_changes.review']);
		expect(requiredPermissionsForPath('/api/users/change-requests/request-1', 'PATCH')).toEqual(['profile_changes.review']);
		expect(requiredPermissionsForPath('/api/users/user-1/reset-password', 'POST')).toEqual(['users.reset_password']);
		expect(requiredPermissionsForPath('/api/users/user-1/profile-link', 'PATCH')).toEqual(['users.update']);
		expect(requiredPermissionsForPath('/api/bank-soal/questions', 'POST')).toEqual(['bank_soal.create']);
		expect(requiredPermissionsForPath('/api/asesmen/packages', 'POST')).toEqual(['asesmen.package_manage']);
		expect(requiredPermissionsForPath('/api/asesmen/events/event-1', 'PATCH')).toEqual(['asesmen.event_manage']);
	});

	it('checks role membership from user payloads', () => {
		expect(hasAnyRole({ id: '1', username: 'guru', role: 'guru', roles: [], permissions: [] }, ['guru'])).toBe(true);
		expect(hasAnyRole({ id: '1', username: 'staf', role: '', roles: ['guru', 'staf'], permissions: [] }, ['admin', 'staf'])).toBe(true);
		expect(hasAnyRole({ id: '1', username: 'ortu', role: 'ortu', roles: [], permissions: [] }, ['admin', 'guru'])).toBe(false);
		expect(hasAnyPermission({ id: '1', username: 'rbac', role: '', roles: [], permissions: [' users.read '] }, ['users.read'])).toBe(true);
	});
});
