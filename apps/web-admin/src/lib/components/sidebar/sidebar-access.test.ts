import { describe, expect, it } from 'vitest';
import { filterSidebarNavGroupsByAccess, hasAnyPermission, itemAllowedByAccess } from './sidebar-access';
import type { SidebarNavGroup } from './sidebar-config';

const groups: SidebarNavGroup[] = [
	{
		group: 'Sistem',
		items: [
			{ href: '/settings', label: 'Pengaturan', icon: 'settings' },
			{ href: '/settings/users', label: 'Manajemen User', icon: 'users', roles: ['admin'], permissions: ['users.read'] },
			{ href: '/settings/audit-logs', label: 'Audit Trail', icon: 'file-text', roles: ['admin'], permissions: ['audit.read'] }
		]
	}
];

describe('sidebar permission access', () => {
	it('allows navigation items by permissions before falling back to legacy roles', () => {
		expect(itemAllowedByAccess(groups[0].items[1], [], ['users.read'])).toBe(true);
		expect(itemAllowedByAccess(groups[0].items[1], ['admin'], [])).toBe(true);
		expect(itemAllowedByAccess(groups[0].items[1], ['guru'], [])).toBe(false);
	});

	it('filters groups using permissions and keeps public items', () => {
		const visible = filterSidebarNavGroupsByAccess(groups, [], ['audit.read']);

		expect(visible).toHaveLength(1);
		expect(visible[0].items.map((item) => item.href)).toEqual(['/settings', '/settings/audit-logs']);
	});

	it('normalizes permission membership checks', () => {
		expect(hasAnyPermission([' users.read ', 'audit.read'], ['users.manage_roles', 'users.read'])).toBe(true);
		expect(hasAnyPermission(['users.read'], ['roles.manage'])).toBe(false);
	});
});
