import { describe, expect, it, vi } from 'vitest';
import {
	buildProfileLinkPayload,
	buildUserRolePayload,
	fetchRBACMatrix,
	resetUserPassword,
	updateUserProfileLink,
	updateUserRoles
} from './rbac-users';

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
});
