import { describe, expect, it } from 'vitest';

import { hasAnyRole, isAdminOnlyPath, isKesiswaanPath, isPublicPath, isStaffOperationPath } from './route-access';

describe('route access helpers', () => {
	it('keeps settings root available to authenticated non-admin users', () => {
		expect(isAdminOnlyPath('/settings')).toBe(false);
		expect(isAdminOnlyPath('/settings/users')).toBe(true);
	});

	it('matches public paths and prefixes', () => {
		expect(isPublicPath('/')).toBe(true);
		expect(isPublicPath('/berita/arsip-kegiatan')).toBe(true);
		expect(isPublicPath('/dashboard')).toBe(false);
	});

	it('matches staff and kesiswaan operation paths', () => {
		expect(isStaffOperationPath('/library/loans')).toBe(true);
		expect(isStaffOperationPath('/api/inventory/items')).toBe(true);
		expect(isKesiswaanPath('/kesiswaan')).toBe(true);
		expect(isKesiswaanPath('/api/kesiswaan/students')).toBe(true);
	});

	it('checks role membership from user payloads', () => {
		expect(hasAnyRole({ id: '1', username: 'guru', role: 'guru', roles: [] }, ['guru'])).toBe(true);
		expect(hasAnyRole({ id: '1', username: 'staf', role: '', roles: ['guru', 'staf'] }, ['admin', 'staf'])).toBe(true);
		expect(hasAnyRole({ id: '1', username: 'ortu', role: 'ortu', roles: [] }, ['admin', 'guru'])).toBe(false);
	});
});
