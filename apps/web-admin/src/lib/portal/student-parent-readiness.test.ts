import { describe, expect, it } from 'vitest';
import readinessDoc from '../../../../../docs/student-parent-account-portal-readiness.md?raw';
import parentsSql from '../../../../../services/core-api/db/queries/parents.sql?raw';
import profileChangeSql from '../../../../../services/core-api/db/queries/profile_change_requests.sql?raw';
import parentPortalService from '../../../../../services/core-api/internal/service/parent_portal.go?raw';
import routeAccess from '../server/route-access.ts?raw';

describe('student parent portal readiness gate', () => {
	it('documents rollout, scope, first-login, change-request, and rollback contracts', () => {
		for (const phrase of [
			'Role and Permission Contract',
			'Account Generation Policy',
			'Username and Password Policy',
			'Student Self-Scope Rules',
			'Parent Child-Scope Rules',
			'Change Request Policy',
			'Smoke Checklist',
			'Rollback Plan',
			'parent_students',
			'must_change_password',
			'target_student_id'
		]) {
			expect(readinessDoc).toContain(phrase);
		}
	});

	it('locks parent child-scope and first-login enforcement in source', () => {
		expect(parentsSql).toContain('GetPortalParentIDByUserID');
		expect(parentsSql).toContain('GetParentPortalChildAccess');
		expect(parentsSql).toContain('parent_scope.parent_id = sqlc.arg(parent_id)');
		expect(parentPortalService).toContain('GetPortalParentIDByUserID');
		expect(parentPortalService).toContain('ensureChildAccess');
		expect(routeAccess).toContain('isMustChangePasswordAllowedPath');
		expect(routeAccess).toContain('/settings/account');
		expect(routeAccess).toContain('/api/auth/change-password');
	});

	it('keeps official data changes on the approval workflow', () => {
		expect(profileChangeSql).toContain('CreateProfileChangeRequest');
		expect(profileChangeSql).toContain('GetParentOwnedChildOfficialProfile');
		expect(profileChangeSql).toContain('JOIN parent_students ps ON ps.parent_id = u.parent_id');
		expect(profileChangeSql).toContain('UpdateStudentOfficialAddress');
		expect(profileChangeSql).toContain('UpdateParentOfficialOccupation');
		expect(profileChangeSql).not.toContain('UPDATE students SET nisn');
	});
});
