import { describe, expect, it } from 'vitest';
import { filterSidebarNavGroupsByAccess, hasAnyPermission, itemAllowedByAccess } from './sidebar-access';
import type { SidebarNavGroup, SidebarNavItem } from './sidebar-config';
import { flattenSidebarNavGroups } from './sidebar-tree';

const accountItem: SidebarNavItem = { href: '/settings/account', label: 'Akun Saya', icon: 'user-check', permissions: ['settings.account'], allowAuthenticatedFallback: true };
const userItem: SidebarNavItem = { href: '/settings/users', label: 'Manajemen User', icon: 'users', roles: ['admin'], permissions: ['users.read'] };
const auditItem: SidebarNavItem = { href: '/settings/audit-logs', label: 'Audit Trail', icon: 'file-text', roles: ['admin'], permissions: ['audit.read'] };
const portalItem: SidebarNavItem = { href: '/portal/siswa', label: 'Portal Siswa', icon: 'book-open', roles: ['siswa'], roleFallbacks: ['siswa'], permissions: ['student_portal.read'] };

const groups: SidebarNavGroup[] = [
	{
		group: 'Sistem',
		items: [
			{
				kind: 'folder',
				id: 'akun-akses',
				href: '',
				label: 'Akun & Akses',
				icon: 'settings',
				permissions: [],
				children: [accountItem, userItem, auditItem]
			},
			{
				kind: 'folder',
				id: 'portal',
				href: '',
				label: 'Portal',
				icon: 'book-open',
				permissions: [],
				children: [portalItem]
			}
		]
	}
];

const visibleHrefs = (visibleGroups: SidebarNavGroup[]) => flattenSidebarNavGroups(visibleGroups).map((item) => item.href);

describe('sidebar permission access', () => {
	it('allows navigation items by permissions while preserving admin access', () => {
		expect(itemAllowedByAccess(userItem, [], ['users.read'])).toBe(true);
		expect(itemAllowedByAccess(userItem, ['admin'], [])).toBe(true);
		expect(itemAllowedByAccess(userItem, ['guru'], [])).toBe(false);
	});

	it('keeps only authenticated basic items when permissions are absent inside nested folders', () => {
		const visible = filterSidebarNavGroupsByAccess(groups, ['guru'], []);

		expect(visible).toHaveLength(1);
		expect(visibleHrefs(visible)).toEqual(['/settings/account']);
	});

	it('keeps explicit portal role fallback separate from generic guru fallback', () => {
		expect(itemAllowedByAccess(portalItem, ['siswa'], [])).toBe(true);
		expect(itemAllowedByAccess(portalItem, ['guru'], [])).toBe(false);
	});

	it('filters nested groups using permissions and keeps account basics', () => {
		const visible = filterSidebarNavGroupsByAccess(groups, [], ['audit.read']);

		expect(visible).toHaveLength(1);
		expect(visibleHrefs(visible)).toEqual(['/settings/account', '/settings/audit-logs']);
	});

	it('normalizes permission membership checks', () => {
		expect(hasAnyPermission([' users.read ', 'audit.read'], ['users.manage_roles', 'users.read'])).toBe(true);
		expect(hasAnyPermission(['users.read'], ['roles.manage'])).toBe(false);
	});
});
