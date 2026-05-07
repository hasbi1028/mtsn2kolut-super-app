import { describe, expect, it } from 'vitest';
import pageSource from './+page.svelte?raw';

describe('/settings/users RBAC scope', () => {
	it('keeps user-role assignment but delegates RBAC catalog editing to /settings/rbac', () => {
		expect(pageSource).toContain('href="/settings/rbac"');
		expect(pageSource).toContain('Manajemen RBAC');
		expect(pageSource).toContain('updateUserRoles');
		expect(pageSource).not.toContain('createRBACRole');
		expect(pageSource).not.toContain('createRBACPermission');
		expect(pageSource).not.toContain('updateRBACRole');
		expect(pageSource).not.toContain('updateRBACPermission');
		expect(pageSource).not.toContain('setRBACRoleActive');
		expect(pageSource).not.toContain('setRBACPermissionActive');
		expect(pageSource).not.toContain("window.prompt('Kode role baru");
		expect(pageSource).not.toContain("window.prompt('Kode permission baru");
	});
});
