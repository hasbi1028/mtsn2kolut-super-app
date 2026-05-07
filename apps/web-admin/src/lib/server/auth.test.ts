import { describe, expect, it } from 'vitest';

import { getUserFromToken } from './auth';

function token(payload: Record<string, unknown>): string {
	const encoded = Buffer.from(JSON.stringify(payload)).toString('base64url');
	return `header.${encoded}.signature`;
}

describe('getUserFromToken', () => {
	it('returns RBAC permissions from the access token', () => {
		const user = getUserFromToken(
			token({
				type: 'access',
				uid: 'user-1',
				usr: 'admin',
				role: 'admin',
				roles: ['admin'],
				permissions: ['users.read', 'roles.manage']
			})
		);

		expect(user?.permissions).toEqual(['users.read', 'roles.manage']);
	});

	it('returns the first-login password change flag from the access token', () => {
		const user = getUserFromToken(
			token({
				type: 'access',
				uid: 'user-1',
				usr: 'siswa001',
				role: 'siswa',
				roles: ['siswa'],
				must_change_password: true
			})
		);

		expect(user?.must_change_password).toBe(true);
	});
});
