import { describe, expect, it } from 'vitest';
import { filterSidebarNavGroupsByAccess, hasAnyPermission, itemAllowedByAccess } from './sidebar-access';
import type { SidebarNavGroup } from './sidebar-config';

const groups: SidebarNavGroup[] = [
	{
		group: 'Sistem',
		items: [
			{ href: '/settings/account', label: 'Akun Saya', icon: 'user-check', permissions: ['settings.account'], allowAuthenticatedFallback: true },
			{ href: '/settings/users', label: 'Manajemen User', icon: 'users', roles: ['admin'], permissions: ['users.read'] },
			{ href: '/settings/audit-logs', label: 'Audit Trail', icon: 'file-text', roles: ['admin'], permissions: ['audit.read'] },
			{ href: '/portal/siswa', label: 'Portal Siswa', icon: 'book-open', roles: ['siswa'], roleFallbacks: ['siswa'], permissions: ['student_portal.read'] }
		]
	}
];

describe('sidebar permission access', () => {
	it('allows navigation items by permissions while preserving admin access', () => {
		expect(itemAllowedByAccess(groups[0].items[1], [], ['users.read'])).toBe(true);
		expect(itemAllowedByAccess(groups[0].items[1], ['admin'], [])).toBe(true);
		expect(itemAllowedByAccess(groups[0].items[1], ['guru'], [])).toBe(false);
	});

	it('keeps only authenticated basic items when permissions are absent', () => {
		const visible = filterSidebarNavGroupsByAccess(groups, ['guru'], []);

		expect(visible).toHaveLength(1);
		expect(visible[0].items.map((item) => item.href)).toEqual(['/settings/account']);
	});

	it('keeps explicit portal role fallback separate from generic guru fallback', () => {
		expect(itemAllowedByAccess(groups[0].items[3], ['siswa'], [])).toBe(true);
		expect(itemAllowedByAccess(groups[0].items[3], ['guru'], [])).toBe(false);
	});

	it('filters groups using permissions and keeps account basics', () => {
		const visible = filterSidebarNavGroupsByAccess(groups, [], ['audit.read']);

		expect(visible).toHaveLength(1);
		expect(visible[0].items.map((item) => item.href)).toEqual(['/settings/account', '/settings/audit-logs']);
	});

	it('normalizes permission membership checks', () => {
		expect(hasAnyPermission([' users.read ', 'audit.read'], ['users.manage_roles', 'users.read'])).toBe(true);
		expect(hasAnyPermission(['users.read'], ['roles.manage'])).toBe(false);
	});
});
