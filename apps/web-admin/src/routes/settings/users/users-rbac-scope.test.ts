import { describe, expect, it } from 'vitest';
import pageSource from './+page.svelte?raw';

describe('/settings/users RBAC scope', () => {
	it('keeps user-role assignment but delegates RBAC catalog editing to /settings/rbac', () => {
		expect(pageSource).toContain('href="/settings/rbac"');
		expect(pageSource).toContain('Manajemen Hak Akses');
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

	it('uses role-scoped profile candidates instead of eager all-profile pulls', () => {
		expect(pageSource).toContain('fetchUserProfileCandidates');
		expect(pageSource).toContain('loadProfileCandidates');
		expect(pageSource).toContain('candidateRequiresClass');
		expect(pageSource).not.toContain("fetch('/api/employees')");
		expect(pageSource).not.toContain("fetch('/api/students')");
		expect(pageSource).not.toContain("fetch('/api/parents')");
	});

	it('surfaces class-aware profile picker copy for siswa and orang tua roles', () => {
		expect(pageSource).toContain('Siswa per kelas');
		expect(pageSource).toContain('Ortu per kelas anak');
		expect(pageSource).toContain('Pilih kelas terlebih dahulu untuk menarik siswa');
		expect(pageSource).toContain('Pilih kelas anak untuk menarik orang tua/wali terkait');
	});
});
