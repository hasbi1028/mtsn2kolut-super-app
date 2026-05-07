import { describe, expect, it } from 'vitest';
import readinessContent from '../../../../../docs/rbac-management-readiness.md?raw';

describe('RBAC management release readiness document', () => {
	it('locks the final RBAC management contract and smoke checklist', () => {
		const requiredPhrases = [
			'/settings/rbac',
			'/settings/users',
			'PUT /api/rbac/roles/{code}/permissions',
			'roles.manage',
			'users.manage_roles',
			'users.manage',
			'permissions.manage',
			'rbac.manage',
			'Session invalidation',
			'Last active admin',
			'401',
			'302',
			'Phase 8',
		];

		for (const phrase of requiredPhrases) {
			expect(readinessContent).toContain(phrase);
		}
	});
});
