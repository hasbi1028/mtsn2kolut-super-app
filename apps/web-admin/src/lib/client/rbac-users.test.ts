import { describe, expect, it, vi } from 'vitest';
import {
	buildProfileLinkPayload,
	buildUserRolePayload,
	fetchRBACMatrix,
	employeeAccountGenerationCSV,
	generateEmployeeAccounts,
	previewEmployeeAccountGeneration,
	createRBACPermission,
	createRBACRole,
	resetUserPassword,
	setRBACPermissionActive,
	setRBACRoleActive,
	updateRBACPermission,
	updateRBACRole,
	updateRBACRolePermissions,
	updateUserProfileLink,
	updateUserRoles
} from './rbac-users';
import { diffPermissions, isCriticalPermission, permissionsByModule, rolePermissionMap } from '$lib/rbac/matrix';

describe('rbac user management client helpers', () => {
	it('normalizes dynamic role payloads for PATCH /api/users/{id}/roles', async () => {
		const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: { ok: true } }), { status: 200 }));

		await updateUserRoles('user-1', ['guru', 'admin', 'guru', ''], fetcher);

		expect(fetcher).toHaveBeenCalledWith('/api/users/user-1/roles', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ roles: ['guru', 'admin'] })
		});
		expect(buildUserRolePayload(['', 'staf', 'staf'])).toEqual({ roles: ['staf'] });
	});

	it('sends reset password and profile link payloads to backend hardening endpoints', async () => {
		const fetcher = vi
			.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }));

		await resetUserPassword('user-1', 'newSecret123', fetcher);
		await updateUserProfileLink('user-1', { employee_id: 'emp-1', student_id: '', parent_id: null }, fetcher);

		expect(fetcher).toHaveBeenNthCalledWith(1, '/api/users/user-1/reset-password', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ password: 'newSecret123' })
		});
		expect(fetcher).toHaveBeenNthCalledWith(2, '/api/users/user-1/profile-link', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ employee_id: 'emp-1', student_id: null, parent_id: null })
		});
		expect(buildProfileLinkPayload({ employee_id: '', student_id: 'stu-1', parent_id: undefined })).toEqual({
			employee_id: null,
			student_id: 'stu-1',
			parent_id: null
		});
	});

	it('fetches RBAC matrix and exposes roles/permissions arrays', async () => {
		const matrix = {
			roles: [{ code: 'admin', name: 'Administrator' }],
			permissions: [{ code: 'users.read', name: 'Baca User' }],
			role_permissions: { admin: ['users.read'] }
		};
		const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: matrix }), { status: 200 }));

		await expect(fetchRBACMatrix(fetcher)).resolves.toEqual(matrix);
		expect(fetcher).toHaveBeenCalledWith('/api/rbac/matrix');
	});

	it('sends CRUD role and permission requests to RBAC management endpoints', async () => {
		const fetcher = vi
			.fn()
			.mockImplementation(() => Promise.resolve(new Response(JSON.stringify({ data: { ok: true } }), { status: 200 })));

		await createRBACRole({ code: 'operator_cbt', name: 'Operator CBT', description: 'Kelola CBT' }, fetcher);
		await updateRBACRole('operator_cbt', { name: 'Operator Asesmen', description: 'Kelola Asesmen' }, fetcher);
		await setRBACRoleActive('operator_cbt', false, fetcher);
		await updateRBACRolePermissions('operator_cbt', ['bank_soal.read', 'bank_soal.review', 'bank_soal.read'], fetcher);
		await createRBACPermission({ code: 'reports.view', module: 'reports', action: 'view', description: 'Lihat laporan' }, fetcher);
		await updateRBACPermission('reports.view', { module: 'reports', action: 'read', description: 'Baca laporan' }, fetcher);
		await setRBACPermissionActive('reports.view', false, fetcher);

		expect(fetcher).toHaveBeenNthCalledWith(1, '/api/rbac/roles', expect.objectContaining({ method: 'POST' }));
		expect(fetcher).toHaveBeenNthCalledWith(2, '/api/rbac/roles/operator_cbt', expect.objectContaining({ method: 'PUT' }));
		expect(fetcher).toHaveBeenNthCalledWith(3, '/api/rbac/roles/operator_cbt/status', expect.objectContaining({ method: 'PATCH', body: JSON.stringify({ is_active: false }) }));
		expect(fetcher).toHaveBeenNthCalledWith(4, '/api/rbac/roles/operator_cbt/permissions', expect.objectContaining({ method: 'PUT', body: JSON.stringify({ permissions: ['bank_soal.read', 'bank_soal.review'] }) }));
		expect(fetcher).toHaveBeenNthCalledWith(5, '/api/rbac/permissions', expect.objectContaining({ method: 'POST' }));
		expect(fetcher).toHaveBeenNthCalledWith(6, '/api/rbac/permissions/reports.view', expect.objectContaining({ method: 'PUT' }));
		expect(fetcher).toHaveBeenNthCalledWith(7, '/api/rbac/permissions/reports.view/status', expect.objectContaining({ method: 'PATCH', body: JSON.stringify({ is_active: false }) }));
	});

	it('normalizes role-permission rows into map and computes matrix diffs', () => {
		const map = rolePermissionMap({
			roles: [{ code: 'guru' }, { code: 'reviewer' }],
			role_permissions: [
				{ role_code: 'guru', permission_code: 'bank_soal.read' },
				{ role_code: 'guru', permission_code: 'bank_soal.read' },
				{ role_code: 'guru', permission_code: 'bank_soal.create' }
			]
		});
		expect(map).toEqual({ guru: ['bank_soal.create', 'bank_soal.read'], reviewer: [] });
		expect(diffPermissions(['bank_soal.read'], ['bank_soal.read', 'bank_soal.publish'])).toEqual({
			added: ['bank_soal.publish'],
			removed: []
		});
		expect(permissionsByModule([
			{ code: 'users.read', module: 'users' },
			{ code: 'bank_soal.read', module: 'bank_soal' },
			{ code: 'misc.view' }
		])).toEqual({
			bank_soal: [{ code: 'bank_soal.read', module: 'bank_soal' }],
			misc: [{ code: 'misc.view' }],
			users: [{ code: 'users.read', module: 'users' }]
		});
		expect(isCriticalPermission('roles.manage')).toBe(true);
		expect(isCriticalPermission('bank_soal.read')).toBe(false);
	});

	it('previews/generates employee accounts and exports initial password CSV', async () => {
		const result = {
			npsn: '40406031',
			default_role: 'guru',
			password_same_as_username: true,
			total: 1,
			ready: 1,
			created: 0,
			skipped: 0,
			failed: 0,
			items: [
				{ employee_id: 'emp-1', nip: '198', nama: 'Hasbi Awal', tanggal_lahir: '1992-10-28', nomor_urut: 1, username: '40406031281092001', password: '40406031281092001', role: 'guru', status: 'ready' as const, message: 'siap dibuat' }
			]
		};
		const fetcher = vi
			.fn()
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: result }), { status: 200 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: { ...result, created: 1, ready: 0 } }), { status: 200 }));

		await expect(previewEmployeeAccountGeneration(fetcher)).resolves.toEqual(result);
		await expect(generateEmployeeAccounts(fetcher)).resolves.toMatchObject({ created: 1, ready: 0 });
		expect(fetcher).toHaveBeenNthCalledWith(1, '/api/users/generate-from-employees/preview');
		expect(fetcher).toHaveBeenNthCalledWith(2, '/api/users/generate-from-employees', { method: 'POST' });
		expect(employeeAccountGenerationCSV(result)).toContain('"Hasbi Awal","198","1992-10-28","40406031281092001","40406031281092001","guru"');
	});

});
