import { describe, expect, it } from 'vitest';

import { hasAnyRole, isAdminOnlyPath, isGuruSafeCbtSupportReadPath, isKesiswaanPath, isPublicPath, isReadMethod, isStaffOperationPath, isStudentApiPath, isStudentPagePath } from './route-access';

describe('route access helpers', () => {
	it('keeps settings root available to authenticated non-admin users', () => {
		expect(isAdminOnlyPath('/settings')).toBe(false);
		expect(isAdminOnlyPath('/settings/users')).toBe(true);
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
		expect(isAdminOnlyPath('/cbt/events')).toBe(true);
		expect(isAdminOnlyPath('/cbt/events/event-1')).toBe(true);
		expect(isAdminOnlyPath('/api/cbt/events')).toBe(true);
		expect(isAdminOnlyPath('/api/cbt/events/event-1')).toBe(true);
		expect(isAdminOnlyPath('/cbt/packages')).toBe(true);
		expect(isAdminOnlyPath('/cbt/sessions')).toBe(true);
		expect(isAdminOnlyPath('/api/cbt/packages')).toBe(true);
		expect(isAdminOnlyPath('/api/cbt/sessions/session-1')).toBe(true);
		expect(isAdminOnlyPath('/api/cbt/questions')).toBe(false);
		expect(isAdminOnlyPath('/cbt/soal')).toBe(false);
		expect(isAdminOnlyPath('/cbt/soal/review')).toBe(false);
		expect(isAdminOnlyPath('/cbt/events/event-1/members')).toBe(true);
		expect(isAdminOnlyPath('/api/school-profile')).toBe(true);
		expect(isAdminOnlyPath('/parentship')).toBe(false);
		expect(isAdminOnlyPath('/api/parentship')).toBe(false);
	});

	it('keeps CBT smoke route boundaries explicit for admin and guru access', () => {
		expect(isPublicPath('/cbt/soal')).toBe(false);
		expect(isAdminOnlyPath('/cbt/soal')).toBe(false);
		expect(isAdminOnlyPath('/cbt/soal/template')).toBe(false);
		expect(isAdminOnlyPath('/api/cbt/questions')).toBe(false);
		expect(isAdminOnlyPath('/api/cbt/questions/export')).toBe(false);
		expect(isAdminOnlyPath('/cbt/events')).toBe(true);
		expect(isAdminOnlyPath('/cbt/events/event-1/exam-cards')).toBe(true);
		expect(isAdminOnlyPath('/api/cbt/events/event-1')).toBe(true);
		expect(isAdminOnlyPath('/api/cbt/soal-support/subjects')).toBe(false);
		expect(isGuruSafeCbtSupportReadPath('/api/cbt/events', 'GET')).toBe(true);
		expect(isGuruSafeCbtSupportReadPath('/api/cbt/events', 'POST')).toBe(false);
		expect(isGuruSafeCbtSupportReadPath('/api/cbt/events/event-1', 'GET')).toBe(false);
		expect(isGuruSafeCbtSupportReadPath('/api/cbt/soal-support/subjects', 'GET')).toBe(true);
		expect(isGuruSafeCbtSupportReadPath('/api/cbt/events/event-1/question-targets', 'GET')).toBe(true);
		expect(isGuruSafeCbtSupportReadPath('/api/cbt/events/event-1/question-targets', 'PUT')).toBe(false);
		expect(isAdminOnlyPath('/cbt/packages')).toBe(true);
		expect(isAdminOnlyPath('/cbt/sessions')).toBe(true);
		expect(isAdminOnlyPath('/cbt/eventship')).toBe(false);
		expect(isAdminOnlyPath('/api/cbt/eventship')).toBe(false);
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

	it('checks role membership from user payloads', () => {
		expect(hasAnyRole({ id: '1', username: 'guru', role: 'guru', roles: [] }, ['guru'])).toBe(true);
		expect(hasAnyRole({ id: '1', username: 'staf', role: '', roles: ['guru', 'staf'] }, ['admin', 'staf'])).toBe(true);
		expect(hasAnyRole({ id: '1', username: 'ortu', role: 'ortu', roles: [] }, ['admin', 'guru'])).toBe(false);
	});
});
